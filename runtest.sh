# open serial port
stty raw -F $1 $2 -echo

exec 3<$1                               # redirect serial to fd3
  cat <&3 > /tmp/testoutput.txt &       # redirect fd3 serial to file
  sleep $4
  PID=$!                                # save pid to kill test if needed
    echo "t" > $1                       # send command to start tests
    sleep $3                            # wait for response
  kill $PID                             # kill test if takes too long
  wait $PID 2>/dev/null                 # surpress "Terminated" output

exec 3<&-                               # free fd3
cat /tmp/testoutput.txt                 # display test results
