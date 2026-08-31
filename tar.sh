#! /bin/sh
filename=`date +%Y%m%d-%H%M`
tar zcvf srMission_$filename.tar.gz \
*.enc.yml \
Titan48272812_cookies \
srMission \
run_srmission.sh \
*.go \
go.mod \
go.sum \
tar.sh