#!/usr/bin/perl

#		AnError(hdlr, www, req, 400, 8001, "Invalid token.")
#
#If SendStatusOnerror is true then,  Status = 400, Internal server error.
#else Status = 200 and JSON response is:
#
#	{"status":"error","code":"9103","msg":"UserName can not be a UUID","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
#	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}

%fruit = (
	'400' => 'Bad Request',
	'401' => 'Unauthorized',
	'402' => 'Payment Required',
	'403' => 'Forbidden',
	'404' => 'Not Found',
	'405' => 'Method Not Allowed',
	'406' => 'Not Acceptable',
	'407' => 'Proxy Auth Required',
	'408' => 'Request Timeout',
	'409' => 'Conflict',
	'410' => 'Gone',
	'411' => 'Length Required',
	'412' => 'Precondition Failed',
	'413' => 'Request Entity Too Large',
	'414' => 'Request U R I Too Long',
	'415' => 'Unsupported Media Type',
	'416' => 'Requested Range Not Satisfiable',
	'417' => 'Expectation Failed',
	'418' => 'Teapot',
	'428' => 'Precondition Required',
	'429' => 'Too Many Requests',
	'431' => 'Request Header Fields Too Large',
	'451' => 'Unavailable For Legal Reasons',
	'500' => 'Internal Server Error',
	'501' => 'Not Implemented',
	'502' => 'Bad Gateway',
	'503' => 'Service Unavailable',
	'504' => 'Gateway Timeout',
	'505' => 'HTTP Version Not Supported',
	'511' => 'Network Authentication Required',
);

while ( <> ) {
	chomp;
	# print "input: $_\n";
	s/^[ 	]*AnError\(hdlr, www, req, //;
	$status_code = $_;
	$status_code =~ s/,.*//;
	s/^[0-9]*, //;
	$error_code = $_;
	$error_code =~ s/,.*//;
	s/^[0-9]*, //;
	$rest = $_;
	$rest =~ s/^"//;
	$rest =~ s/"\).*//;
	$lookkup_status_code = $fruit{$status_code};
	# print " status_code= >$status_code< error_code= >$error_code< rest= >$rest<\n";
	$title = $rest;
	$title =~ tr/A-Z/a-z/;
	print <<XXxx

##### $title

If SendStatusOnerror is true then,  Status = $status_code, $lookkup_status_code.
else Status = 200 and JSON response is:

	{"status":"error","code":"$error_code","msg":"$rest","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}

XXxx
}

exit (0);

__END__

Examples of Input:
		AnError(hdlr, www, req, 400, 8004, "Invalid salt.")
		AnError(hdlr, www, req, 400, 8004, "Invalid salt.")
		AnError(hdlr, www, req, 400, 8004, "Invalid salt.")
		AnError(hdlr, www, req, 400, 8004, "Invalid salt.")
		AnError(hdlr, www, req, 400, 8004, "Invalid salt.")
		AnError(hdlr, www, req, 400, 8004, "Invalid salt.")
		AnError(hdlr, www, req, 400, 8005, "Invalid 'v' verifier value.")
		AnError(hdlr, www, req, 400, 8005, "Invalid 'v' verifier value.")
		AnError(hdlr, www, req, 400, 8005, "Invalid 'v' verifier value.")
		AnError(hdlr, www, req, 400, 8005, "Invalid 'v' verifier value.")
		AnError(hdlr, www, req, 400, 8005, "Invalid 'v' verifier value.")
		AnError(hdlr, www, req, 400, 8005, "Invalid 'v' verifier value.")

