package parseCertSpec

import (
	"crypto/tls"
	"fmt"
	"testing"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
// https://gist.github.com/cyx/55d0faf5df60daedae44 -- example of server and client in go.
/*
You can supply all of that information on the command line.

One step self signed passwordless certificate generation:

openssl req \
    -new \
    -newkey rsa:4096 \
    -days 365 \
    -nodes \
    -x509 \
    -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com" \
    -keyout www.example.com.key \
    -out www.example.com.cert
All of the openssl subcommands have their own man page. See man req.

Specifically addressing your questions and to be more explicit about exactly which options are in effect:

The -nodes flag signals to not encrypt the key, thus you do not need a password. You could also use the -passout arg flag. See PASS PHRASE ARGUMENTS in the openssl(1) man page for how to format the arg.

Using the -subj flag you can specify the subject (example is above).
*/

func Test_Pcs1(t *testing.T) {

	tests := []struct {
		spec     string
		expected []uint16
	}{
		{"ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:ECDHE-RSA-DES-CBC3-SHA:ECDHE-ECDSA-DES-CBC3-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA",
			[]uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			}},
		{"ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!3DES:!MD5:!PSK",
			[]uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			}},
	}

	for ii, vv := range tests {
		rv := Pcs(vv.spec)
		fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> [%d] rv=%s\n", ii, StringSliceUInt16(rv))
		//if vv.expectedId != rv.Id {
		if false {
			t.Errorf("Error %2d, Expeced ID = %d, found %d\n", ii, vv.expected, rv)
		}
	}

}

func StringSliceUInt16(a []uint16) (rv string) {
	rv += "{"
	com := ""
	for _, vv := range a {
		rv += com + fmt.Sprintf("%04x", vv)
		com = ", "
	}
	rv += "}"
	return
}

func CompareSlice(a, b []uint16) (ok bool) {
	// TODO
	return
}
