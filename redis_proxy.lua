#!/usr/bin/env luvit
-- -*- mode: lua; tab-width: 2; indent-tabs-mode: 1; st-rulers: [70] -*-
-- vim: ts=4 sw=4 ft=lua noet
---------------------------------------------------------------------
-- @author Daniel Barney <daniel@pagodabox.com>
-- @copyright 2014, Pagoda Box, Inc.
-- @doc
--
-- @end
-- Created :   17 Oct 2014 by Daniel Barney <daniel@pagodabox.com>
---------------------------------------------------------------------
local JSON  = require("json")
local fs = require('fs')

local Main = require('./lib/main.lua')

-- set up default location for the config file
local configPath = "/opt/local/etc/redundis/redundis.conf"

-- it can be specified with the first parameter to the command
if process.argv[1] then
	configPath = process.argv[1]
end


-- load config file
fs.readFile(configPath,function(err,data)

	local main

	-- the config file is completely optional
  if not err then
    local opts = JSON.parse(data)
    main = Main:new(opts)
  else
    main = Main:new()
  end

  main:check_master()

  process:on('error', function(err)
    p("global error: ",{err=err})
  end)

end)
