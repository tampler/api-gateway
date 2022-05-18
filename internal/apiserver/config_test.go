package apiserver

const (
	timeFormat = "02 Jan 06 15:04:05 MST"
	userID     = "12eb8d3e-ea8a-4aa1-9226-5d3762aa668e"
	portStart  = 8085
	portEnd    = 9085
	sleepTime  = 200

	testZone   = "Sandbox-simulator"
	testDomain = "ROOT"
	testAcc    = "admin"

	//ssh
	sshKeyName = "bku-ssh"
	pubkey     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDRXZk6v4lDkTkuVHnx/Ztuqv6ntlc6ry5cLjRGyRKOuPGyyaWkK5I1Y2/vtsK8FV6VOJ0Hdjz63kCNaNHtTieDq8W8q2yL2OYiUrgb4cQf3nPs185i41twZBEG12sCBGoXoYNoJl0WsysZ4SlHPgXF+W8BaQK8aJZmFc/f2upjgzX5HxTNhPV5e2ttpvGisH/r8jJBlLZclQa4DHyhq1iTJWNz7DJq6jh4VxqagriRYabuDJRPtTYpi8v5t6+jWbggGIqQkliSaSyYzpHBZAn4PHWUZdRME738IOI2Jy831DH0wvJ0KVjBlcvrT3yXc92iQ9z0s6tFpuQrxMVL3J9+3NmLtKf4i8dcJWDospiQBJp8DrWEVybV34tJk2nHPVzJFpYgJW2XqXdDQhUmQP9CH6L57IDi5Z4vyFvDtcgFd5PFCvkqA7s0PAMF7PY6+laN45qQiO02NFWQHPXbdFyxjzhsHAJPWGWCuPJMwk16fdRgnodk+Ut7j4AfYxSlyRk= bku@lap"

	// VPC
	vpcName    = "test-vpc"
	updVpcName = "upd-test-vpc"
	vpcCidr4   = "10.0.1.0/24"
	emptyVPCID = ""
	vpcOffer   = "defaultVPC"

	// Network
	netOffer   = "privateVPC"
	netName    = "net-vpc-0"
	netCidr4   = "10.0.1.0/30"
	netCidr6   = "2002::1234:abcd:ffff:c0a8:101/64"
	emptyCIDR6 = ""
	netDomain  = "my.local"
	updNetName = "upd-foo"
	updNetCidr = "10.0.1.0/28"

	// Template
	tmplName        = "My Ubuntu"
	updTemplateName = "My Upd Ubuntu"
	osOffer         = "Ubuntu 21.04"
	tmplFilter      = "all"
	tmplURL         = "http://dl.openvm.eu/cloudstack/macchinina/x86_64/macchinina-xen.vhd.bz2"

	// Instance
	instName    = "ubuntu-vm0"
	updInstName = "upd-vm0"
	diskSizeGB  = 1

	instOffer     = "Small Instance"
	instDiskOffer = "Small"

	// ACL
	aclName  = "my-acl"
	aclDescr = "Description for my-acl"
)
