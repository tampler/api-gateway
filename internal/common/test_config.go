package common

const (
	TimeFormat = "02 Jan 06 15:04:05 MST"
	PortStart  = 8085
	PortEnd    = 9085
	SleepTime  = 200

	TestZone   = "Sandbox-simulator"
	TestDomain = "ROOT"
	TestAcc    = "admin"

	//ssh
	SshKeyName = "bku-ssh"
	Pubkey     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDRXZk6v4lDkTkuVHnx/Ztuqv6ntlc6ry5cLjRGyRKOuPGyyaWkK5I1Y2/vtsK8FV6VOJ0Hdjz63kCNaNHtTieDq8W8q2yL2OYiUrgb4cQf3nPs185i41twZBEG12sCBGoXoYNoJl0WsysZ4SlHPgXF+W8BaQK8aJZmFc/f2upjgzX5HxTNhPV5e2ttpvGisH/r8jJBlLZclQa4DHyhq1iTJWNz7DJq6jh4VxqagriRYabuDJRPtTYpi8v5t6+jWbggGIqQkliSaSyYzpHBZAn4PHWUZdRME738IOI2Jy831DH0wvJ0KVjBlcvrT3yXc92iQ9z0s6tFpuQrxMVL3J9+3NmLtKf4i8dcJWDospiQBJp8DrWEVybV34tJk2nHPVzJFpYgJW2XqXdDQhUmQP9CH6L57IDi5Z4vyFvDtcgFd5PFCvkqA7s0PAMF7PY6+laN45qQiO02NFWQHPXbdFyxjzhsHAJPWGWCuPJMwk16fdRgnodk+Ut7j4AfYxSlyRk= bku@lap"

	// VPC
	VpcName    = "test-vpc"
	UpdVpcName = "upd-test-vpc"
	VpcCidr4   = "10.0.1.0/24"
	EmptyVPCID = ""
	VpcOffer   = "defaultVPC"

	// Network
	NetOffer   = "privateVPC"
	NetName    = "net-vpc-0"
	NetCidr4   = "10.0.1.0/30"
	NetCidr6   = "2002::1234:abcd:ffff:c0a8:101/64"
	EmptyCIDR6 = ""
	NetDomain  = "my.local"
	UpdNetName = "upd-foo"
	UpdNetCidr = "10.0.1.0/28"

	// Template
	TmplName        = "My Ubuntu"
	UpdTemplateName = "My Upd Ubuntu"
	OsOffer         = "Ubuntu 21.04"
	TmplFilter      = "all"
	TmplURL         = "http://dl.openvm.eu/cloudstack/macchinina/x86_64/macchinina-xen.vhd.bz2"

	// Instance
	InstName    = "ubuntu-vm0"
	UpdInstName = "upd-vm0"
	DiskSizeGB  = 1

	InstOffer     = "Small Instance"
	InstDiskOffer = "Small"

	// ACL
	AclName  = "my-acl"
	AclDescr = "Description for my-acl"

	// ACL Rule
	AclrDesc      = "Allow dummy subnet"
	AclrAction    = "allow"
	AclrProto     = "tcp"
	AclrTraffic   = "ingress"
	AclrCIDR4     = "10.100.1.0/24,10.100.2.0/24"
	AclrPortStart = 9991
	AclrPortEnd   = 9995
)
