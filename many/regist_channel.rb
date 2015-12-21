#!/usr/bin/env ruby
#encoding=utf-8
require 'redis'
# redis = Redis.new(:url => "redis://192.168.100.231:6379")
redis = Redis.new(:url => "redis://:abcd@192.168.100.236:6379")
conf = ""
redis.set "channel:fake_teding_recv", "teding_topic"

count = ARGV.first.to_i || 100
count.times do |i|
	c = "fake_recv_#{i}"
  key = "channel:#{c}"
  value = "topicabcdefg:#{c}"
  redis.set key, value
  conf += "#{c}=/u/#{value}\n"
end

File.open "channel.list", "w" do |f|
  f.write conf
end
