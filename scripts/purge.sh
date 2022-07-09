# This script purges all items in the NATs cache and queues

#!/usr/bin/bash

ajc task purge
ajc queue purge PING
ajc queue purge PONG