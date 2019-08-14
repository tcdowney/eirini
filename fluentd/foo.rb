#!/usr/bin/env ruby
# -*- coding: utf-8 -*-

RubyVM::InstructionSequence.compile_option = {trace_instruction: false} rescue nil

here = File.dirname(__FILE__)
$LOAD_PATH << File.expand_path(File.join(here, '..', 'lib'))

require 'memory_profiler'

MemoryProfiler.start

at_exit do
  report = MemoryProfiler.stop
  report.pretty_print(to_file: "/var/log/fluentd_memory_profile.#{Process.pid}")
end

require 'fluent/command/fluentd'
