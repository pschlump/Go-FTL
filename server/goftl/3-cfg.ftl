{
    "http://localhost:3111": { "LineNo":2,
		"path_prefix": [
			{ "LineNo":4,
				"list": [ "/api/special/" ],
				"middleware": [
					"log_it":true,
					"dumpIt":"Just after htmlMod",
					"htmlMangle":true,
					"dumpIt":"Just after proxy",
					"proxy_to":[ "http://localhost:8080/{{.URI}}" ]
				]
			},
			{ "LineNo":4,
				"list": [ "/api/", "/xyz/" ],
				"middleware": [
					"log_it":true,
					"dumpIt":"Just after htmlMod",
					"htmlMangle":true,
					"dumpIt":"Just after proxy",
					"proxy_to":[ "http://localhost:8204/{{.URI}}" ]
				]
			},
			{ "LineNo":14,
				"list": [ "/" ],
				"middleware": [
					"log_it":true,
					"dumpIt":"Just after file",
					"static_dir":[ "./www" ]
				]
			}
		]
    }
}
