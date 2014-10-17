local Emitter = require('core').Emitter
local net = require('net')
local string = require('string')
local timer = require('timer')
local redis = require('../deps/luvit-redis')


local Main = Emitter:extend()

function Main:initialize()
	self.enabled = false
	self.current_master = nil
	self.sentinal = redis:new('127.0.0.1',26379)
	self.sentinal:on('error',function(err)
		p(err)
	end)
	self:on('master',self.handle_master)
	self:on('refresh',self.verify_master)
end

function Main:check_master()
	self.sentinal.redisNativeClient:command('SENTINEL','get-master-addr-by-name','test',function(err,dest,...)
		if err then
			p('unable to talk to sentinal: ',err)
			timer.setTimeout(5000,process.exit,1)
			return
		end
		if not dest then
			p('Local sentinal instance is not ready.')
			timer.setTimeout(5000,process.exit,1)
			return
		end

		-- convert to an integer
		dest[2] = 0 + dest[2]
		self:emit('master',self,dest)
		timer.setTimeout(1000,self.check_master,self)
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
		if not self.current_master then
			p('starting up socket server')
			local main = self
			self.server = net.createServer(function(client)
				main:handle_connection(client)
			end)
			self.server:listen(2000)
		else 
			p('refresing client connections')
		end
		self.current_master = master
		self:emit('refresh',self)
	end
end

function Main:verify_master()
	local con = self.current_master
	local check = redis:new(con[1],con[2])
	check:on('error',function() end)
	check.redisNativeClient:command('INFO','replication',function(err,res)
		check:disconnect()
		if err then
			p('unable to verify master')
			timer.setTimeout(1500,self.verify_master,self)
		else
			local role = string.match(res,'role:master')
			if not role then
				p('master not ready yet')
				timer.setTimeout(1000,self.verify_master,self)
			elseif role then
				p('master is ready at ',con[1],con[2])
				self.enabled = true
			end
		end
	end)
end


function Main:handle_connection(client)
	if self.enabled then
		client:resume()
		local con = self.current_master
		local server
		server = net.createConnection(con[2],con[1])

		p('piping connection')

		client:pipe(server)
		server:pipe(client)

		self:once('refresh',function()
			client:destroy()
		end)
	else
		client:pause()
		p('holding connection until it is ready')
		timer.setTimeout(1000,self.handle_connection,self,client)
	end
end

local start = Main:new()

process:on('error', function(err)
	p(err)
end)

start:check_master()

return 1