# Readme
Process:
1. ./amcli org add -email test@test.org
Successfully created org OrgID: 3 UserID: 3
2. ./amcli group add -input s3://home/fbi/gohome/src/github.com/linkai-io/am/cmd/amcli/hosts.txt -oid 3 -uid 3 -name test
Successfully created new scangroup: oid: 3 gid: 1
3. ./amcli addr add -input hosts.txt -oid 3 -uid 3 -gid 1
Successfully added 10 addresses for OrgID: 3
4. ./amcli coor start -gid 1 -oid 3 -uid 3
Successfully started scangroup: oid: 3 gid: 1