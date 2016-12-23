TODO Documentation
==================

Added to config

	ApiTableKey      string                      // If true (!= "") then this password will be requried to access /api/table calls.

Added to respond to this

			if hdlr.ApiTableKey != "" {
				ps := &rw.Ps
				pwSupplied := ps.ByNameDflt("api_table_key", "")
				if hdlr.ApiTableKey != pwSupplied {
					if !hdlr.final || hdlr.Next == nil {
						trx.AddNote(1, "TabServer2: final - return - 406")
						logrus.Errorf("406 api_table_key did not match required key: %s", godebug.LF())
						www.WriteHeader(http.StatusNotAcceptable)
					} else {
						hdlr.Next.ServeHTTP(www, req)
					}
					return
				}
			}

Key is "api_table_key" -- this is the param that must be suppled if used.

* Must test
* Add in prod

