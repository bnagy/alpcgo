require 'buggery'
require 'thread'
require 'bindata'
require 'hexdump'

# http://www.retrojunkie.com/asciiart/animals/pandas.htm
# ASCII art credit - Normand Veilleux
panda = <<eos

         d8888 888      8888888b.   .d8888b.  888      .d88888b.   .d8888b.  
        d88888 888      888   Y88b d88P  Y88b 888     d88P" "Y88b d88P  Y88b 
       d88P888 888      888    888 888    888 888     888     888 888    888 
      d88P 888 888      888   d88P 888        888     888     888 888        
     d88P  888 888      8888888P"  888        888     888     888 888  88888 
    d88P   888 888      888        888    888 888     888     888 888    888 
   d8888888888 888      888        Y88b  d88P 888     Y88b. .d88P Y88b  d88P 
  d88P     888 88888888 888         "Y8888P"  88888888 "Y88888P"   "Y8888P88 

                              (c) @rantyben 2014

                               ~ SPONSORED BY ~


                              _,add8ba,
                            ,d888888888b,
                           d8888888888888b                        _,ad8ba,_
                          d888888888888888)                     ,d888888888b,
                          I8888888888888888 _________          ,8888888888888b
                __________`Y88888888888888P"""""""""""baaa,__ ,888888888888888,
            ,adP"""""""""""9888888888P""^                 ^""Y8888888888888888I
         ,a8"^           ,d888P"888P^                           ^"Y8888888888P'
       ,a8^            ,d8888'         _____________________       ^Y8888888P'
      a88'           ,d8888P'         [ UNIT 61398 RESERVE  ]         I88P"^
    ,d88'           d88888P'          [  THOMAS "CB" LIM    ]         "b,
   ,d88'           d888888'           [ CALLSIGN SEXY PANDA ]          `b,
  ,d88'           d888888I             `````````````````````            `b,
  d88I           ,8888888'            ___                                `b,
 ,888'           d8888888          ,d88888b,              ____            `b,
 d888           ,8888888I         d88888888b,           ,d8888b,           `b
,8888           I8888888I        d8888888888I          ,88888888b           8,
I8888           88888888b       d88888888888'          8888888888b          8I
d8886           888888888       Y888888888P'           Y8888888888,        ,8b
88888b          I88888888b      `Y8888888^             `Y888888888I        d88,
Y88888b         `888888888b,      `""""^                `Y8888888P'       d888I
`888888b         88888888888b,                           `Y8888P^        d88888
 Y888888b       ,8888888888888ba,_          _______        `""^        ,d888888
 I8888888b,    ,888888888888888888ba,_     d88888888b               ,ad8888888I
 `888888888b,  I8888888888888888888888b,    ^"Y888P"^      ____.,ad88888888888I
  88888888888b,`888888888888888888888888b,     ""      ad888888888888888888888'
  8888888888888698888888888888888888888888b_,ad88ba,_,d88888888888888888888888
  88888888888888888888888888888888888888888b,`"""^ d8888888888888888888888888I
  8888888888888888888888888888888888888888888baaad888888888888888888888888888'
  Y8888888888888888888888888888888888888888888888888888888888888888888888888P
  I888888888888888888888888888888888888888888888P^  ^Y8888888888888888888888'
  `Y88888888888888888P88888888888888888888888888'     ^88888888888888888888I
   `Y8888888888888888 `8888888888888888888888888       8888888888888888888P'
    `Y888888888888888  `888888888888888888888888,     ,888888888888888888P'
     `Y88888888888888b  `88888888888888888888888I     I888888888888888888'
       "Y8888888888888b  `8888888888888888888888I     I88888888888888888'
         "Y88888888888P   `888888888888888888888b     d8888888888888888'
            ^""""""""^     `Y88888888888888888888,    888888888888888P'
                            "8888888888888888888b,   Y888888888888P^
                             `Y888888888888888888b   `Y8888888P"^
                                "Y8888888888888888P     `""""^
                                  `"YY88888888888P'
                                       ^""""""""'
eos

target = ARGV[0]
fail "Usage: #{$0} <target pid>" unless Integer(target)

debugger = Buggery::Debugger.new

# I untangled a log of unions here, see ntlpcapi.h for details
class PORT_MESSAGE < BinData::Record
  endian :little
  uint16 :data_length
  uint16 :total_length
  uint16 :type
  uint16 :data_info_offset
  uint64 :process
  uint64 :thread
  uint32 :message_id
  uint32 :pad
  uint64 :client_view_size # or callback id
end

PORT_MESSAGE_SIZE = 0x28

# Do parsing and display in a separate thread so that the callback proc can be
# as speedy as possible
q = Queue.new
Thread.new do
  loop do

    s = q.pop
    m = PORT_MESSAGE.read(s)

    puts '='*80
    puts
    puts "Type:     0x%x" % m.type
    puts "Process:  #{m.process}"
    puts "Id:       #{m.message_id}"
    puts
    puts Hexdump.dump s[PORT_MESSAGE_SIZE..-1]
    puts

  end
end

bp_proc = lambda {|args|

  begin

    p_msg = debugger.read_pointers( debugger.registers['rsp']+0x28 ).first
    return 1 if p_msg.null?

    # hackily get total length
    message_offset = p_msg.address
    total_length = debugger.read_virtual( message_offset+2, 2 ).unpack('s').first

    if total_length >= PORT_MESSAGE_SIZE
      message = debugger.read_virtual(message_offset, total_length)
      q.push message
    end

  ensure
    return 1 # DEBUG_STATUS_GO
  end

}

debugger.event_callbacks.add breakpoint: bp_proc

begin
  debugger.attach target
  debugger.break
  debugger.wait_for_event # post attach
rescue
  fail "Unable to attach: #{$!}\n#{$@.join("\n")}"
end

# ntdll!ZwAlpcSendWaitReceivePort:
# 00000000`77041b60 4c8bd1          mov     r10,rcx
# 00000000`77041b63 b882000000      mov     eax,82h
# 00000000`77041b68 0f05            syscall
# 00000000`77041b6a c3              ret <--- BREAK
#
# We break after the syscall, which is when the kernel has filled in the
# receive buffer
debugger.execute "bp ntdll!NtAlpcSendWaitReceivePort+0xa"

puts panda
puts "Breakpoint set, starting processing loop."
puts "Hit ^C to exit...\n\n"

# This seems convoluted, but JRuby is kind of weird about threads, so it's
# best to be extra nice to it.
abort = Queue.new
Signal.trap "INT" do
  abort.push true
end

debugger.go

loop do

  begin

    debugger.wait_for_event(200) # ms

    break unless debugger.has_target?

    unless abort.empty?
      puts "Caught abort, trying a clean exit..."
      debugger.detach_process
      puts "Detatched!"
      break
    end

  rescue

    puts "Caught error, exiting: #{$!}"
    break

  end

end
