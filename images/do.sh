#!/bin/bash

chmod 777 /tmp/in.txt
(while read LINE
do 
./code << EOF 
    $LINE 
EOF
done  < /tmp/in.txt
) > /tmp/out.txt

