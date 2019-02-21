#!/usr/bin/perl

my $DOC="./srp_aes_auth.md", $line_no = 0;

while ( <> ) {
	$line_no++;
	if ( /^func / ) {
		print "$line_no: $_";
	}
}


