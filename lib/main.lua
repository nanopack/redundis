local Emitter = require('core').Emitter
local net = require('net')
local string = require('string')
local timer = require('timer')
local redis = require('../deps/luvit-redis')


local Main = Emitter:extend()

function Main:initialize(opts)
	self.enabled = false
	self.current_master = nil
	self.opts = opts
	self:normalize_opts()
	self.sentinel = redis:new(self.opts.sentinel_ip,self.opts.sentinel_port)

	self.sentinel:on('error',function(err)
		p(err)
	end)

	-- we start up the server socket
	self.server = net.createServer(function(client)
		self:handle_connection(client)
	end)
	self.server:listen(self.opts.listen_port,self.opts.listen_ip)

end

function add_default(key,value,object)
	if not object[key] then
		object[key] = value 
	end
end

function Main:normalize_opts()
	if not self.opts then
		self.opts = {}
	end

	local opts = self.opts
	add_default('sentinel_ip','127.0.0.1',opts)
	add_default('sentinel_port',26379,opts)
	add_default('listen_ip','127.0.0.1',opts)
	add_default('listen_port',6379,opts)

	add_default('monitor_name','test',opts)
	add_default('not_ready_timeout',5000,opts)
	add_default('sentinel_poll_timeout',1000,opts)
	add_default('master_wait_timeout',1000,opts)
end

function Main:check_master()
	-- SENTINEL is not a supported command, so we bypass the builtin stuff
	self.sentinel.redisNativeClient:command('SENTINEL','get-master-addr-by-name',self.opts.monitor_name,function(err,dest,...)
		if err then
			p('unable to talk to sentinel: ',err)
			timer.setTimeout(self.opts.not_ready_timeout,self.check_master,self)
			return
		end
		if not dest then
			p('Local sentinel instance is not ready.')
			timer.setTimeout(self.opts.not_ready_timeout,self.check_master,self)
			return
		end

		-- convert port to an integer
		dest[2] = 0 + dest[2]
		self:handle_master(dest)
		timer.setTimeout(self.opts.sentinel_poll_timeout,self.check_master,self)
	end)
end

function Main:need_refresh(master)
	if not self.current_master then
		return true
	else
		return not (master[1] == self.current_master[1]) or not (master[2] == self.current_master[2])
	end
end

function Main:handle_master(master)
	if self:need_refresh(master) then
		p('refresing client connections')

		-- we disable all future connections
		self.enabled = false
		
		self.current_master = master
		self:emit('begin-refresh',self)
		self:verify_master()
	end
end

function Main:verify_master()
	local con = self.current_master
	local check = redis:new(con[1],con[2])

	-- we need to check if the new master has become a master node yet
	check:on('error',function() end)

	-- we use INFO instead of ROLE because ROLE is not a supported 
	-- command in all versions of redis
	check.redisNativeClient:command('INFO','replication',function(err,res)
		check:disconnect()
		if err then
			p('unable to verify master')
			timer.setTimeout(self.opts.master_wait_timeout,self.verify_master,self)
		else
			local role = string.match(res,'role:master')
			if not role then
				p('master not ready yet')
				timer.setTimeout(self.opts.master_wait_timeout,self.verify_master,self)
			elseif role then
				p('master is ready at ',con[1],con[2])

				-- we enable all future connections
				self.enabled = true
				self:emit('refresh',self)
			end
		end
	end)
end


function Main:handle_connection(client)
	if self.enabled then
		local con = self.current_master
		local server
		server = net.createConnection(con[2],con[1])

		-- just pipe it all
		client:pipe(server)
		server:pipe(client)

		-- once the failover has happened, we close all connections
		self:once('begin-refresh',function()
			client:destroy()
		end)

	else

		-- if we are currently waiting for a failover to complete
		-- pause all connections and wait for the refresh to happen

		client:pause()
		p('holding connection until it is ready')
		self:once('refresh',function()
			client:resume()
			self:handle_connection(client)
		end)

	end
end

return Main