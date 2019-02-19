package TabServer2

//
// R E S T s e r v e r - Server Component	(TabServer2)
//
// Copyright (C) Philip Schlump, 2012-2017 -- All rights reserved.
//
// Do not remove the following lines - used in auto-update.
// Version: 1.1.0
// BuildNo: 0391
// FileId: 0005
// File: TabServer2/crud.go
//

// xyzzy-JWT

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug" //	"encoding/json"
	"github.com/pschlump/ms"
	"github.com/pschlump/uuid"
)

//	"github.com/pschlump/Authorize.Net"
// ====================================================================================================================================================================
// --------------------------------------------------------------------- Table Ops ------------------------------------------------------------------------------------
// ====================================================================================================================================================================

func (hdlr *TabServer2Type) validOp(name string, op string, cfgTag string) bool {
	h := hdlr.SQLCfg[cfgTag] // get configuration						// Xyzzy - what if config not defined for this item at this point!!!!
	// fmt.Printf ( "len=%d inArray=%v h.Crud=%v op=%s cfgTag=[%s] h=%v\n", len(h.Crud), sizlib.InArray(op,h.Crud), h.Crud, op, cfgTag, h )
	if len(h.Crud) > 0 && sizlib.InArray(op, h.Crud) {
		return true
	}
	// fmt.Printf ( "Invalid Operation: len=%d inArray=%v h.Crud=%v op=%s cfgTag=[%s] h=%v\n", len(h.Crud), sizlib.InArray(op,h.Crud), h.Crud, op, cfgTag, h )
	return false
}

type WhereClause struct {
	Op       string
	Val1s    string
	Val2s    string
	Val1i    int64
	Val2i    int64
	Val1f    float64
	Val2f    float64
	Val1d    time.Time
	Val2d    time.Time
	Name     string
	List     []WhereClause
	Expr     []WhereClause
	CastName string
}

func AddBindValue(bind *[]interface{}, x interface{}) (pos int) {
	pos = len(*bind) + 1
	*bind = append(*bind, x)
	return
}

func ValidateColInWhere(name string, h SQLOne) (ty string, err error) {
	for _, v := range h.Cols {
		if v.ColName == name {
			return v.ColType, nil // may want to check to see if this is a legit col for a where clause?
		}
	}
	return "", errors.New("Unable to find column")
}

// Used in "in" or "not in" in where clause - where data is to be bound to a list of values.
func (hdlr *TabServer2Type) GetDataList(ty string, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}) (string, error) {
	s := "("
	com := ""
	bp := 0
	for _, vv := range wc.List {
		switch ty {
		case "i":
			bp = AddBindValue(bind, vv.Val1i)
			s += com + hdlr.BindPlaceholder(bp)
		case "f":
			bp = AddBindValue(bind, vv.Val1f)
			s += com + hdlr.BindPlaceholder(bp)
		case "u": // UUID/GUID
			// Odbc-xyzzy - if ODBC - then put in a conversion
			bp = AddBindValue(bind, vv.Val1s)
			if hdlr.GetDbType() == DbType_postgres {
				s += com + hdlr.BindPlaceholder(bp)
			} else if hdlr.GetDbType() == DbType_odbc {
				s += com + " convert(UniqueIdentifier," + hdlr.BindPlaceholder(bp) + ") "
			} else if hdlr.GetDbType() == DbType_Oracle {
				s += com + hdlr.BindPlaceholder(bp)
			} else {
				s += com + hdlr.BindPlaceholder(bp)
			}
		case "":
			fallthrough
		case "s":
			bp = AddBindValue(bind, vv.Val1s)
			s += com + hdlr.BindPlaceholder(bp)
		case "d":
			bp = AddBindValue(bind, vv.Val1d)
			s += com + hdlr.BindPlaceholder(bp)
		/* d, t, e */
		default:
			trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s", ty, wc.Name))
			return "", errors.New(fmt.Sprintf("Error(10033): Invalid Type: %s for column %s", ty, wc.Name))
		}
		com = ","
	}
	s = s + ")"
	return s, nil
}

// The where clause is passed as a parse tree, 'wc'.   This is to be translated into a corresponding string and returned.
// This somewhat limits what can be passed as a where clause.   This also prevents SQL injection.   If you need a more
// complex where clause then a stored procedure can be written that is specific to the task at hand.
func GenWhereFromWc(wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}, hdlr *TabServer2Type) (string, error) {
	bp := 0
	b2 := 0
	s := ""
	if wc.Op == "and" || wc.Op == "or" {
		and := ""
		for _, vv := range wc.List {
			if sizlib.InArray(vv.Op, []string{"==", "!=", "<>", ">=", "<=", ">", "<", "=", "like", "not like"}) {
				fmt.Printf("Found == or other similar, %s\n", godebug.LF())
				// xyzzy-m5 - vv.Name is name of column - check to see if this is a "attr" for this object or "key_word" or "category"
				pp, ss, err := ExtendeAttributes(vv, wc, trx, h, bind, hdlr)
				if err != nil {
					fmt.Printf("At, %s, %s\n", godebug.LF(), err)
					trx.AddNote(1, fmt.Sprintf("Error: %v processing extended attributes %s%s", err, h.setWhereAlias, vv.Name))
					return "", errors.New(fmt.Sprintf("Error(14239): Error: %v processing extended attributes %s%s", err, h.setWhereAlias, vv.Name))
				} else if pp {
					fmt.Printf("At, %s ->%s<-\n", godebug.LF(), ss)
					s += and + ss
				} else {
					fmt.Printf("%sAt, %s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
					ty, err := ValidateColInWhere(vv.Name, h)
					if (vv.Op == "like" || vv.Op == "not like") && (ty == "i" || ty == "f" || ty == "d") {
						trx.AddNote(1, fmt.Sprintf("Genrally not a good idea to use like/not-like with a numeric or date type; for column %s%s", h.setWhereAlias, vv.Name))
					}
					if err != nil {
						trx.AddNote(1, fmt.Sprintf("Type Error: %v for column %s%s", err, h.setWhereAlias, vv.Name))
						return "", errors.New(fmt.Sprintf("Error(10039): Type Error: %v for column %s%s", err, h.setWhereAlias, vv.Name))
					} else {
						switch ty {
						case "i":
							bp = AddBindValue(bind, vv.Val1i)
							s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						case "f":
							bp = AddBindValue(bind, vv.Val1f)
							s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						case "u": // UUID/GUID
							// Odbc-xyzzy -- Definitly convert in data from string -> UniqueIdentifier
							bp = AddBindValue(bind, vv.Val1s)
							if hdlr.GetDbType() == DbType_postgres {
								s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
							} else if hdlr.GetDbType() == DbType_odbc {
								s += and + fmt.Sprintf(`%s%s%s%s %s convert(UniqueIdentifier,`+hdlr.BindPlaceholder(bp)+`)`, h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
							} else if hdlr.GetDbType() == DbType_Oracle {
								s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
							} else {
								s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
							}
						case "":
							fallthrough
						case "s":
							bp = AddBindValue(bind, vv.Val1s)
							s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						case "d":
							if len(vv.Expr) > 0 {
								t, err := GenWhereFromWc(vv.Expr[0], trx, h, bind, hdlr)
								if err != nil {
									return "", err
								}
								// crud.go:1798: missing argument for Sprintf("%s"): format reads arg 5, have only 4 args
								// ORIG:
								// s += and + fmt.Sprintf(`%s%s%s%s %s %s `, h.setWhereAlias, vv.Name, vv.Op, t)
								s += and + fmt.Sprintf(`%s%s%s%s %s %s`, h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op, t)
							} else {
								bp = AddBindValue(bind, vv.Val1d)
								s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
							}
						/* d, t, e */
						default:
							trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s%s", ty, h.setWhereAlias, vv.Name))
							return "", errors.New(fmt.Sprintf("Error(10035): Invalid Type: %s for column %s%s%s%s", ty, h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
						}
					}
				}
			} else if sizlib.InArray(vv.Op, []string{"-", "+", "*", "/"}) {
				s += fmt.Sprintf(`%s%s%s %s %s '%s' `, DbBeginQuote, vv.Name, DbEndQuote, vv.Op, vv.CastName, vv.Val1s)
				// current_timestamp - interval '122 days'
				// if vv.Name is a function/constant
			} else if sizlib.InArray(vv.Op, []string{"between", "not between"}) {
				ty, err := ValidateColInWhere(vv.Name, h)
				if err != nil {
					trx.AddNote(1, fmt.Sprintf("Type Error: %v for column %s%s%s%s", err, h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
					return "", errors.New(fmt.Sprintf("Error(10038): Type Error: %v for column %s%s%s%s", err, h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
				} else {
					switch ty {
					case "i":
						bp = AddBindValue(bind, vv.Val1i)
						b2 = AddBindValue(bind, vv.Val1i)
						s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, vv.Name, vv.Op)
					case "f":
						bp = AddBindValue(bind, vv.Val1f)
						b2 = AddBindValue(bind, vv.Val2f)
						s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
					case "u": // UUID/GUID
						// Odbc-xyzzy -- Definitly convert in data from string -> UniqueIdentifier
						bp = AddBindValue(bind, vv.Val1s)
						b2 = AddBindValue(bind, vv.Val2s)
						if hdlr.GetDbType() == DbType_postgres {
							s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						} else if hdlr.GetDbType() == DbType_odbc {
							s += and + fmt.Sprintf(`%s%s%s%s %s convert(UniqueIdentifier,`+hdlr.BindPlaceholder(bp)+`) and convert(UniqueIdentifier,`+hdlr.BindPlaceholder(b2)+`)`,
								h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						} else if hdlr.GetDbType() == DbType_Oracle {
							s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						} else {
							s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
						}
					case "":
						fallthrough
					case "s":
						bp = AddBindValue(bind, vv.Val1s)
						b2 = AddBindValue(bind, vv.Val2s)
						s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
					case "d":
						bp = AddBindValue(bind, vv.Val1d)
						b2 = AddBindValue(bind, vv.Val2d)
						s += and + fmt.Sprintf(`%s%s%s%s %s `+hdlr.BindPlaceholder(bp)+` and `+hdlr.BindPlaceholder(b2), h.setWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op)
					/* d, t, e */
					default:
						trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s%s%s%s", ty, h.SetWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
						return "", errors.New(fmt.Sprintf("Error(10036): Invalid Type: %s for column %s%s%s%s", ty, h.SetWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
					}
				}
			} else if sizlib.InArray(vv.Op, []string{"in", "not in"}) {
				pp, ss, err := ExtendeAttributes(vv, wc, trx, h, bind, hdlr)
				if err != nil {
					trx.AddNote(1, fmt.Sprintf("Error: %v processing extended attributes %s%s", err, h.setWhereAlias, vv.Name))
					return "", errors.New(fmt.Sprintf("Error(14239): Error: %v processing extended attributes %s%s", err, h.setWhereAlias, vv.Name))
				} else if pp {
					s += and + ss
				} else {
					ty, err := ValidateColInWhere(vv.Name, h)
					if err != nil {
						trx.AddNote(1, fmt.Sprintf("Type Error: %v for column %s%s%s", err, DbBeginQuote, vv.Name, DbEndQuote))
						return "", errors.New(fmt.Sprintf("Error(10037): Type Error: %v for column %s%s%s%s", err, h.SetWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
					} else {
						lst, err := hdlr.GetDataList(ty, vv, trx, h, bind)
						if err != nil {
							return "", err
						}
						s += and + fmt.Sprintf(`%s%s%s%s %s %s`, h.SetWhereAlias, DbBeginQuote, vv.Name, DbEndQuote, vv.Op, lst)
						if ty == "f" {
							trx.AddNote(1, fmt.Sprintf("Genrally not a good idea to use an in list with a float type; for column %s%s%s%s", h.SetWhereAlias, DbBeginQuote, vv.Name, DbEndQuote))
						}
					}
				}
			} else if sizlib.InArray(vv.Op, []string{"and", "or"}) {
				t, err := GenWhereFromWc(vv, trx, h, bind, hdlr)
				if err != nil {
					return "", err
				}
				s += and + "(" + t + ")"
			} else {
				trx.AddNote(1, fmt.Sprintf("Invalid Op: %s", vv.Op))
				return "", errors.New(fmt.Sprintf("Error(10040): Invalid Op: %s", vv.Op))
			}
			// xyzzy - missing regular expressions - PostgreSQL specific ops
			and = " " + wc.Op + " "
		}
	} else {
		return "", errors.New("Error(10041): Must have an 'and' or 'or' as the top level in the where clause")
	}
	return s, nil
}

// Select -
//	1. xyzzy /api/table/NAME/ID - to get a row
//	2. /api/table/NAME?where... - to be more selective
//	3. /api/table/NAME?where...&orderBy=...
//
// Update - general where clauses
// Delete - general where clauses
// ins-upd ops
//
// Xyzzy - Weekness - insert - should be able to return an array of IDs - more than one AutoGen column
// Xyzzy - Weekness - insert - should be able to use sequence for values - and call d.b. for it [ return set of them too ]

// http://localhost:8090/api/table/t_email_q?where={"op":"and","List":[{"op":"=","name":"to","val1s":"pschlump@gmail.com"},{"op":"=","name":"status","val1s":"pending"}]}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) CrudErrMsg(depth int, msg string, err error, col string, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) {
	rv := fmt.Sprintf(`{ "status":"error", "msg":%q, %s }`, msg, godebug.LFj(2+depth))
	trx.SetQryDone(rv, "")
	ReturnErrorMessage(406, "Database Error", "18043",
		fmt.Sprintf(`Error(18043): Database Error (%q) %s`, msg, godebug.LF(2+depth)),
		res, req, *ps, trx, hdlr) // status:error
}

// =========================================================================================================================================================================
/*
 For MS SQL Server:
 		From: https://technet.microsoft.com/en-us/library/gg699618(v=sql.110).aspx
			SELECT First Name + ' ' + Last Name FROM Employees ORDER BY First Name OFFSET 10 ROWS FETCH NEXT 5 ROWS ONLY;

		Test: http://192.168.0.161:8200/api/table/t_test_crud3?limit=3&offset=2
		Test: http://192.168.0.161:8200/api/table/t_test_crud3?limit=3
		Test: http://192.168.0.161:8200/api/table/t_test_crud3?offset=5

 Oracle:

	My all-time-favorite use of ROWNUM is pagination. In this case, I use ROWNUM to get rows N through M of a result set. The general form is as follows:

		select *
		  from ( select / *+ FIRST_ROWS(n) * /
		  a.*, ROWNUM rnum
			  from ( your_query_goes_here,
			  with order by ) a
			  where ROWNUM <=
			  :MAX_ROW_TO_FETCH )
		where rnum  >= :MIN_ROW_TO_FETCH;

	Notes: From: http://www.oracle.com/technetwork/issue-archive/2006/06-sep/o56asktom-086197.html

	Test:
		http://192.168.0.154:8200/api/table/t_test_crud3?limit=3&offset=2
		http://192.168.0.154:8200/api/table/t_test_crud3?limit=3
		http://192.168.0.154:8200/api/table/t_test_crud3?offset=5

*/
// =========================================================================================================================================================================
func (hdlr *TabServer2Type) CommonLimitOffset(mdata map[string]string, data *[]interface{}, h SQLOne, table_name string, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) bool {
	trx.AddNote(2, "In CommonLimitOffset")
	mdata["before_query"] = ""
	mdata["after_query"] = ""
	mdata["limit_limit"] = ""
	mdata["limit"] = ""
	mdata["limit_after"] = ""
	mdata["offset_offset"] = ""
	mdata["offset"] = ""
	mdata["offset_after"] = ""
	var have_limit = false
	var have_offset = false
	var offset = 0
	var limit = 0
	limit_s, ok_limit_s := ps.GetByName("limit")
	offset_s, ok_offset_s := ps.GetByName("offset")
	if ok_limit_s && limit_s != "0" {
		limit, err := strconv.Atoi(limit_s)
		if err != nil {
			trx.AddNote(2, "In CommonLimitOffset: Invalid value for limit")
			hdlr.CrudErrMsg(2, fmt.Sprintf("Error(15001): Invalid value for limit.  At %s", h.LineNo), err, "", trx, res, req, ps)
			return false
		}
		if limit < 0 {
			trx.AddNote(2, "In CommonLimitOffset: Invalid value for limit")
			hdlr.CrudErrMsg(2, fmt.Sprintf("Error(15003): Invalid value for limit.  Must be > 0.  At %s", h.LineNo), nil, "", trx, res, req, ps)
			return false
		}
		trx.AddNote(2, fmt.Sprintf("have limit of %v", limit))
		have_limit = true
	}
	if ok_offset_s && offset_s != "0" {
		offset, err := strconv.Atoi(offset_s)
		if err != nil {
			trx.AddNote(2, "In CommonLimitOffset: Invalid value for offset")
			hdlr.CrudErrMsg(2, fmt.Sprintf("Error(15002): Invalid value for offset.  At %s", h.LineNo), err, "", trx, res, req, ps)
			return false
		}
		if offset < 0 {
			trx.AddNote(2, "In CommonLimitOffset: Invalid value for offset")
			hdlr.CrudErrMsg(2, fmt.Sprintf("Error(15005): Invalid value for offset.  Must be > 0.  At %s", h.LineNo), nil, "", trx, res, req, ps)
			return false
		}
		trx.AddNote(2, fmt.Sprintf("have offset of %v", offset))
		have_offset = true
	}
	if len(h.OrderBy) == 0 && hdlr.GetDbType() == DbType_Oracle {
		trx.AddNote(2, "In CommonLimitOffset: Order by must be specified as a default in sql-cfg.json for Oracle to perform properly.")
		hdlr.CrudErrMsg(2, fmt.Sprintf("Error(15000): Order By must be specified as a default in sql-cfg.json for Oracle to perform properly.  At %s", h.LineNo), nil,
			"", trx, res, req, ps)
		return false
	}
	if hdlr.GetDbType() == DbType_Oracle {
		trx.AddNote(2, "In CommonLimitOffset: Oracle processing query")
		if have_offset && have_limit {
			mdata["before_query"] = fmt.Sprintf("select * from ( select /*+ FIRST_ROWS(%d) */ aaa.*, ROWNUM r___num from ( ", limit)
			mdata["after_query"] = fmt.Sprintf(" ) aaa where ROWNUM <= %d ) where r___num >= %d ", limit, offset)
		} else if have_limit {
			mdata["before_query"] = fmt.Sprintf("select /*+ FIRST_ROWS(%d) */ aaa.* ( ", limit)
			mdata["after_query"] = fmt.Sprintf(" ) aaa where ROWNUM <= %d ", limit)
		} else if have_offset {
			mdata["before_query"] = "select * from ( "
			mdata["after_query"] = fmt.Sprintf(" ) aaa where ROWNUM >= %d ", offset)
		}
	} else {
		if have_limit {
			if hdlr.GetDbType() == DbType_postgres {
				trx.AddNote(2, "In CommonLimitOffset: PostgreSQL processing limit of:"+limit_s)
				mdata["limit_limit"] = "limit"
				mdata["limit"] = limit_s
			} else if hdlr.GetDbType() == DbType_odbc {
				trx.AddNote(2, "In CommonLimitOffset: MS-SQL processing limit of:"+limit_s)
				mdata["limit_limit"] = "fetch next "
				mdata["limit"] = limit_s
				mdata["limit_after"] = " rows only"
				if !have_offset {
					have_offset = true
					offset_s = "0"
				}
			} else {
				fmt.Printf("Not implemented yet, %s\n", godebug.LF())
			}
		}
		if have_offset {
			if hdlr.GetDbType() == DbType_postgres {
				trx.AddNote(2, "In CommonLimitOffset: PostgreSQL processing offset of:"+offset_s)
				mdata["offset_offset"] = "offset"
				mdata["offset"] = offset_s
			} else if hdlr.GetDbType() == DbType_odbc {
				trx.AddNote(2, "In CommonLimitOffset: MS-SQL processing offset of:"+offset_s)
				mdata["offset_offset"] = "offset "
				mdata["offset"] = offset_s
				mdata["offset_after"] = " rows"
			} else {
				fmt.Printf("Not implemented yet, %s\n", godebug.LF())
			}
		}
	}
	return true
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) CommonWhereClause(pkRequired bool, id string, idflag bool, mdata map[string]string, data *[]interface{}, h SQLOne, table_name string, trx *tr.Trx, wc *WhereCollect, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) (fail bool, isPk bool) {
	trx.AddNote(2, "In CommonWhereClause")
	var err error
	if len(h.SetWhereAlias) > 0 && len(h.setWhereAlias) == 0 {
		h.setWhereAlias = `"` + h.SetWhereAlias + `".`
	}
	ty := ""
	n_pk := 0
	bp := 0

	// fmt.Printf("h.Cols=%s, idflag=%v, pkRequired=%v, %s\n", godebug.SVar(h.Cols), idflag, pkRequired, godebug.LF())
	// if sizlib.InArray("dump_insert_params", h.DebugFlag) {

	if len(h.Cols) > 0 {
		s := ""
		com := ""
		pk_found := false
		for ii, v := range h.Cols {
			addWhereClause := false
			if v.IsPk {

				n_pk++
				if n_pk > 1 && idflag {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10043): Only 1 column can be specified as a primary key when using this API. %s%s at %s", h.setWhereAlias, v.ColName, h.LineNo), nil, v.ColName, trx, res, req, ps)
					return false, false
				}
				ty, err = ValidateColInWhere(v.ColName, h)
				if err != nil {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10044): Type Error: %v for column %s%s", err, h.setWhereAlias, v.ColName), err, v.ColName, trx, res, req, ps)
					return false, false
				}

				addWhereClause = true

			} else if v.IsIndexed {

				ty, err = ValidateColInWhere(v.ColName, h)
				if err != nil {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10044): Type Error: %v for column %s%s", err, h.setWhereAlias, v.ColName), err, v.ColName, trx, res, req, ps)
					return false, false
				}

				if sizlib.InArray("dump_general_where", hdlr.DebugFlags) {
					fmt.Printf("Matched column [%s] AT: %s\n", v.ColName, godebug.LF())
				}

				addWhereClause = true

			}

			if addWhereClause {

				switch ty {

				case "u": // UUID/GUID
					// Odbc-xyzzy -- Definitly convert in data from string -> UniqueIdentifier
					if idflag {
						pk_found = true
						bp = AddBindValue(data, id)
						if hdlr.GetDbType() == DbType_postgres {
							s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						} else if hdlr.GetDbType() == DbType_odbc {
							s = s + com + fmt.Sprintf(` %s%s%s%s = convert(UniqueIdentifier,`+hdlr.BindPlaceholder(bp)+`)`, h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						} else if hdlr.GetDbType() == DbType_Oracle {
							s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						}
						com = " and "
					} else if ps.HasName(v.ColName) {
						pk_found = true
						bp = AddBindValue(data, ps.ByName(v.ColName))
						if hdlr.GetDbType() == DbType_postgres {
							s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						} else if hdlr.GetDbType() == DbType_odbc {
							s = s + com + fmt.Sprintf(` %s%s%s%s = convert(UniqueIdentifier,`+hdlr.BindPlaceholder(bp)+`)`, h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						} else if hdlr.GetDbType() == DbType_Oracle {
							s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						}
						com = " and "
					}

				case "":
					fallthrough
				case "s":
					if idflag {
						pk_found = true
						bp = AddBindValue(data, id)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					} else if ps.HasName(v.ColName) {
						pk_found = true
						bp = AddBindValue(data, ps.ByName(v.ColName))
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					}

				case "i":
					// fmt.Printf("At: %s\n", godebug.LF())
					var x int64
					if idflag {
						pk_found = true
						x, err = strconv.ParseInt(id, 10, 64)
						bp = AddBindValue(data, x)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					} else if ps.HasName(v.ColName) {
						pk_found = true
						// fmt.Printf("v.ColName >%s< At: %s\n", v.ColName, godebug.LF())
						x, err = strconv.ParseInt(ps.ByNameDflt(v.ColName, "0"), 10, 64)
						bp = AddBindValue(data, x)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					}
					if err != nil {
						hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10046): Invalid Number column %s, %s%s%s%s, LineNo:%v", err, h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote, h.LineNo), err, v.ColName, trx, res, req, ps)
						return false, false
					}

				case "b":
					var b bool
					if idflag {
						pk_found = true
						b = sizlib.ParseBool(id)
						bp = AddBindValue(data, b)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					} else if ps.HasName(v.ColName) {
						pk_found = true
						b = sizlib.ParseBool(ps.ByNameDflt(v.ColName, "false"))
						bp = AddBindValue(data, b)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					}

				case "f":
					var f float64
					if idflag {
						pk_found = true
						f, err = strconv.ParseFloat(id, 64)
						bp = AddBindValue(data, f)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					} else if ps.HasName(v.ColName) {
						pk_found = true
						f, err = strconv.ParseFloat(ps.ByNameDflt(v.ColName, "0"), 64)
						bp = AddBindValue(data, f)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					}
					if err != nil {
						hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10046): Invalid Number column %s, %s%s%s%s, LineNo:%v", err, h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote, h.LineNo), err, v.ColName, trx, res, req, ps)
						return false, false
					}

				case "d":
					fallthrough
				case "t":
					fallthrough
				case "e":
					var d time.Time
					nullOk := ValidateNullOk(v.ColName, ii, h, "where")
					if idflag {
						pk_found = true
						d, _, err = ms.FuzzyDateTimeParse(id, nullOk)
						bp = AddBindValue(data, d)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					} else if ps.HasName(v.ColName) {
						pk_found = true
						d, _, err = ms.FuzzyDateTimeParse(ps.ByNameDflt(v.ColName, "0"), nullOk)
						bp = AddBindValue(data, d)
						s = s + com + fmt.Sprintf(` %s%s%s%s = `+hdlr.BindPlaceholder(bp), h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote)
						com = " and "
					}
					if err != nil {
						hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10047): Invalid Time/Date column %s, %s%s%s%s, LineNo:%v", err, h.setWhereAlias, DbBeginQuote, v.ColName, DbEndQuote, h.LineNo), err, v.ColName, trx, res, req, ps)
						return false, false
					}
				}
			}
		}
		// fmt.Printf("At: %s\n", godebug.LF())
		ok3 := ps.HasName("where")
		if pk_found {
			// fmt.Printf("s = >>>%s<<< At: %s\n", s, godebug.LF())
			isPk = true
			wc.AddClause(s)
		} else if pkRequired && !ok3 {
			// fmt.Printf("At: %s\n", godebug.LF())
			hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10045): Table must have a primary key when using this API. None Specified. %s at [sql-cfg.json]line_no=%s", table_name, h.LineNo), nil, "", trx, res, req, ps)
			return false, false
		}
	}
	return true, isPk
}

// ====================================================================================================================================================================
// Validate the set of columns passed based on 'h.valid'
// Used in "set" for update, func GenUpdateSet ( h SQLOne, m url.Values, ...
// Used in "insert" values section, func GenInsertValues ( h SQLOne, ...
// ====================================================================================================================================================================
func ValidateSetValue(ColName string, ii int, h SQLOne, dd string, req *http.Request) (err error) {
	// return nil
	/*
	   mm22mm
	   xyzzy-nullOk - why is this commented out - test and verify
	*/

	// If no validation has been set then.... Assume all is good.
	if !HasKeys(h.valid) {
		return nil
	}

	err = nil

	var vv *map[string]Validation // mm22mm
	vv = &h.valid
	switch req.Method {
	case "GET": // Select
		if HasKeys(h.validGet) {
			vv = &h.validGet
		}
	case "POST": // Insert
		if HasKeys(h.validPost) {
			vv = &h.validPost
		}
	case "PUT": // Update
		if HasKeys(h.validPut) {
			vv = &h.validPut
		}
	case "DELETE": // Delete
		if HasKeys(h.validDel) {
			vv = &h.validDel
		}
	default:
		err = errors.New(fmt.Sprintf("Error(14012): Interal error - should never reach this code, %s", godebug.LF()))
		return
	}

L:
	// for i, v := range h.valid {
	for i, v := range *vv {
		if i == ColName {
			// found_it = true
			// fmt.Printf ( "Update|Insert:Validate i=%s, v=%s\n", i, sizlib.SVar(v) )

			switch v.Type {

			case "u": // validation - this is ok
				fallthrough
			case "s":
				if v.eMin_len {
					if len(dd) < v.Min_len {
						err = errors.New(fmt.Sprintf("Error(10082): Parameter (%s) Too Short:%s Minimum Length %d value=[%s]", dd, i, v.Min_len, dd))
						return
					}
				}
				if v.eMax_len {
					if len(dd) > v.Max_len {
						err = errors.New(fmt.Sprintf("Error(10083): Parameter Too Long:%s Maximum Length %d value=[%s]", i, v.Max_len, dd))
						return
					}
				}
				if v.eReMatch {
					matched, err2 := regexp.MatchString(v.ReMatch, dd)
					if err2 != nil {
						err = errors.New(fmt.Sprintf("Error(10084): Rgular expression in valiation invalid - error %s", err2))
						return
					}
					if !matched && v.Required {
						err = errors.New(fmt.Sprintf("Error(10085): Parameter failed to match regular expression:%s", i))
						return
					}
				}

			case "i":
				var w int64
				if v.eMin || v.eMax || v.ChkType {
					w, err = strconv.ParseInt(dd, 10, 64)
					if err != nil {
						err = errors.New(fmt.Sprintf("Error(10086): Invalid integer - failed to parse at %s, %s, value=%s", i, err, dd))
						return
					}
				}
				if v.eMin {
					if w < v.Min {
						// fmt.Printf ( " -- failed min --\n" );
						err = errors.New(fmt.Sprintf("Error(10087): Parameter Too Short:%s Minimum %d", i, v.Min))
						return
					}
				}
				if v.eMax {
					if w > v.Max {
						// fmt.Printf ( " -- failed max --\n" );
						err = errors.New(fmt.Sprintf("Error(10088): Parameter Too Large:%s Maximum %d", i, v.Max))
						return
					}
				}

			case "f":
				var w float64
				if v.eMinF || v.eMaxF || v.ChkType {
					w, err = strconv.ParseFloat(dd, 64)
					if err != nil {
						err = errors.New(fmt.Sprintf("Error(10089): Invalid number - failed to parse at %s, %s, value=%s", i, err, dd))
						return
					}
				}
				if v.eMinF {
					if w < v.MinF {
						// fmt.Printf ( " -- failed min --\n" );
						err = errors.New(fmt.Sprintf("Error(10090): Parameter Too Small:%s Minimum %f", i, v.MinF))
						return
					}
				}
				if v.eMaxF {
					if w > v.MaxF {
						// fmt.Printf ( " -- failed max --\n" );
						err = errors.New(fmt.Sprintf("Error(10091): Parameter Too Large:%s Maximum %f", i, v.MaxF))
						return
					}
				}

			case "d":
				fallthrough
			case "t":
				fallthrough
			case "e":
				var w time.Time
				if v.eMinD || v.eMaxD || v.ChkType {
					// w, err = time.Parse( ISO8601, dd)
					// nullOk := ValidateNullOk ( v.ColName, ii, h, "update" )
					/*
					   xyzzy-nullOk not certain that this is ok.
					*/
					nullOk := true
					w, _, err = ms.FuzzyDateTimeParse(dd, nullOk)
					if err != nil {
						err = errors.New(fmt.Sprintf("Error(10092): Invalid date/time - failed to parse at %s, %s, value=%s", i, err, dd))
						return
					}
				}
				if v.eMinD {
					if w.Before(v.MinD) {
						// fmt.Printf ( " -- failed min date --\n" );
						err = errors.New(fmt.Sprintf("Error(10093): Parameter too far in the past:%s Minimum %v", i, v.MinD))
						return
					}
				}
				if v.eMaxD {
					if w.After(v.MaxD) {
						// fmt.Printf ( " -- failed max date --\n" );
						err = errors.New(fmt.Sprintf("Error(10094): Parameter too far in the future:%s Maximum %v", i, v.MaxD))
						return
					}
				}

			}

			break L
		}
	}
	//if ! found_it {
	//	err = errors.New ( fmt.Sprintf ( "Error(10095): Failed to find information to validate." ) )
	//	return
	//}
	err = nil
	return

}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) GenUpdateSet(h SQLOne, trx *tr.Trx, data *[]interface{}, mdata map[string]string, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) bool {

	var err error
	bp := 0

	n_set := 0

	//
	if db_GenUpdateSet {
		fmt.Printf("GenUpdateSet: from %s, %s\n", godebug.LF(), ps.DumpParamTable())
	}

	if len(h.Cols) > 0 {
		s := " set "
		com := ""
		for ii, v := range h.Cols {
			colName := v.ColName
			if v.Update {

				if v.DataColName != "" {
					colName = v.DataColName
				}

				if db_GenUpdateSet {
					fmt.Printf("h.Cols[%d]=%s, %s\n", ii, sizlib.SVar(v), godebug.LF())
				}

				if val, ok := ps.GetByName(colName); ok {
					if db_GenUpdateSet {
						fmt.Printf("have data for it, %s, %s\n", colName, godebug.LF())
					}
					switch v.ColType {
					case "u":
						// Odbc-xyzzy -- Definitly convert in data from string -> UniqueIdentifier
						bp = AddBindValue(data, val)
						if hdlr.GetDbType() == DbType_postgres {
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						} else if hdlr.GetDbType() == DbType_odbc {
							s = s + com + fmt.Sprintf(` %s%s%s = convert(UniqueIdentifier,`+hdlr.BindPlaceholder(bp)+`)`, DbBeginQuote, v.ColName, DbEndQuote)
						} else if hdlr.GetDbType() == DbType_Oracle {
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						}
						err = ValidateSetValue(v.ColName, ii, h, val, req)
						if err != nil {
							hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10058): Invalid string data %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
							return false
						}
						n_set++

					case "":
						fallthrough
					case "s":
						bp = AddBindValue(data, val)
						s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						err = ValidateSetValue(v.ColName, ii, h, val, req)
						if err != nil {
							hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10058): Invalid string data %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
							return false
						}
						n_set++

					case "i":
						var x int64
						nullOk := ValidateNullOk(v.ColName, ii, h, "update")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` %s%s%s = NULL `, DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							x, err = strconv.ParseInt(ps.ByNameDflt(colName, "0"), 10, 64)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10050): Invalid Number column %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false
							}
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10057): Invalid integer data %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false
							}
							bp = AddBindValue(data, x)
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						}
						n_set++

					case "b":
						var b bool
						// fmt.Printf ( " Is bool - type 'b' " );
						nullOk := ValidateNullOk(v.ColName, ii, h, "update")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` %s%s%s = NULL `, DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							b = sizlib.ParseBool(ps.ByNameDflt(colName, "false"))
							// fmt.Printf ( " value = %v ", b );
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10056): Invalid boolean data %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false
							}
							bp = AddBindValue(data, b)
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						}
						n_set++

					case "f":
						var f float64
						nullOk := ValidateNullOk(v.ColName, ii, h, "update")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` %s%s%s = NULL `, DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							f, err = strconv.ParseFloat(ps.ByNameDflt(colName, "0"), 64)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10052): Invalid Number column %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false
							}
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10059): Invalid float data %s, %s, [sql-cfg.json]line_no=%s", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false
							}
							bp = AddBindValue(data, f)
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						}
						n_set++

					case "d":
						fallthrough
					case "t":
						fallthrough
					case "e":
						var d time.Time
						nullOk := ValidateNullOk(v.ColName, ii, h, "update")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` %s%s%s = NULL `, DbBeginQuote, v.ColName, DbEndQuote)
						} else {
							d, _, err = ms.FuzzyDateTimeParse(val, nullOk)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10053): Invalid Time/Date column %s, %s, [sql-cfg.json]line_no=%s date=%s", err, v.ColName, h.LineNo, ps.ByName(v.ColName)), err, v.ColName, trx, res, req, ps)
								return false
							}
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10055): Invalid Time/Date data %s, %s, [sql-cfg.json]line_no=%s date=%s", err, v.ColName, h.LineNo, ps.ByName(v.ColName)), err, v.ColName, trx, res, req, ps)
								return false
							}
							bp = AddBindValue(data, d)
							s = s + com + fmt.Sprintf(` %s%s%s = `+hdlr.BindPlaceholder(bp), DbBeginQuote, v.ColName, DbEndQuote)
						}
						n_set++

					}
					com = ","
				} else {
					if db_GenUpdateSet {
						fmt.Printf("*** have NO data *** for it, %s, %s\n", colName, godebug.LF())
					}
				}
			} else if v.IsPk {
			} else {
				if ps.HasName(colName) {
					fmt.Printf(`{"type":"not-used", "ColName": "%s", "note":"value supplied for column but not used because not an update-able column in specification", "LineFile":%q }`+"\n", colName, godebug.LF())
					trx.AddNote(1, fmt.Sprintf(`Column [%s] Not Used: value supplied for column but not used because not an update-able column in specification Line:%s`, colName, godebug.LF()))
				}
			}
		}
		if db_GenUpdateSet {
			fmt.Printf("n_set = %d, %s\n", n_set, godebug.LF())
			fmt.Printf("updcols = %s, %s\n", s, godebug.LF())
		}
		mdata["updcols"] = s
	}
	if n_set == 0 {
		hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10054): Update must set someting - nothing set. %s.  Check that you supplied a valid column name to update.", h.LineNo), nil, "", trx, res, req, ps)
		return false
	}
	return true
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func GetRequiredByValidation(name string, h SQLOne, crud_op string) (isRequired bool, hasValid bool) {
	if HasKeys(h.valid) {
		hasValid = true
		isRequired = true

		var vv *map[string]Validation // mm22mm
		vv = &h.valid
		switch crud_op { // req.Method {
		case "insert": // "POST":
			if HasKeys(h.validPost) {
				vv = &h.validPost
			}
		case "update": // "PUT":
			if HasKeys(h.validPut) {
				vv = &h.validPut
			}
		}

		// func isRequired(h SQLOne,name string) bool {
		v, ok := (*vv)[name]
		if !ok {
			isRequired = false
		} else {
			isRequired = v.Required
		}
	}
	return
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
// nullOk := ValidateNullOk ( v.ColName, ii, h, "insert" )
// "key" is for generation of a "key" into caching.  Should return constant false.
// "where" caluse - can not have a "where" x = null situraiton - return false.
func ValidateNullOk(ColName string, ii int, h SQLOne, crud_op string) bool {
	switch crud_op {
	case "insert", "update":
		// xyzzy - should check "valid" and see if requried field.
		if isRequiredFlag, hasValid := GetRequiredByValidation(ColName, h, crud_op); hasValid {
			return !isRequiredFlag
		}
		return true
	case "key", "where":
		return false
	default:
		return true
	}
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) GenInsertValues(h SQLOne, trx *tr.Trx, data *[]interface{}, mdata map[string]string, n_col int, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) (ok bool, s_id string) {
	n_gen := 0
	ok = true
	s_id = ""
	bp := 0
	var err error

	used := make(map[string]bool)

	if len(h.Cols) > 0 {
		s := ""
		com := ""
		for ii, v := range h.Cols {
			if v.Insert {

				// fmt.Printf("col is:%s %s\n", v.ColName, godebug.LF())

				// xyzzy - if v.DataColName != "", then ps.GetByName ( v.DataColName ) else ...
				colName := v.ColName
				if v.DataColName != "" {
					colName = v.DataColName
				}
				used[colName] = true

				// fmt.Printf("colName is:%s %s\n", colName, godebug.LF())

				if val, ok := ps.GetByName(colName); ok {

					// fmt.Printf("found colName is:%s %s\n", colName, godebug.LF())

					switch v.ColType {
					case "u": // UUID, GUID
						// Odbc-xyzzy -- Definitly convert in data from string -> UniqueIdentifier
						bp = AddBindValue(data, val)
						if hdlr.GetDbType() == DbType_postgres {
							s = s + com + ` ` + hdlr.BindPlaceholder(bp)
						} else if hdlr.GetDbType() == DbType_odbc {
							s = s + com + ` convert(UniqueIdentifier,` + hdlr.BindPlaceholder(bp) + `)`
						} else if hdlr.GetDbType() == DbType_Oracle {
							s = s + com + ` ` + hdlr.BindPlaceholder(bp)
						} else {
							s = s + com + ` ` + hdlr.BindPlaceholder(bp)
						}
						err = ValidateSetValue(v.ColName, ii, h, val, req)
						if err != nil {
							hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10060): Invalid string data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
							return false, ""
						}
						n_gen++

					case "": // Not Specified, assume it is a string
						fallthrough
					case "s": // string
						bp = AddBindValue(data, val)
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
						err = ValidateSetValue(v.ColName, ii, h, val, req)
						if err != nil {
							hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10060): Invalid string data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
							return false, ""
						}
						n_gen++

					case "i": // Integer
						var x int64
						nullOk := ValidateNullOk(v.ColName, ii, h, "insert")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` NULL `)
						} else {
							x, err = strconv.ParseInt(ps.ByNameDflt(colName, "0"), 10, 64)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10061): Invalid Number column %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false, ""
							}
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10062): Invalid integer data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false, ""
							}
							bp = AddBindValue(data, x)
							s = s + com + " " + hdlr.BindPlaceholder(bp)
						}
						n_gen++

					case "b": // Boolean
						var b bool
						nullOk := ValidateNullOk(v.ColName, ii, h, "insert")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` NULL `)
						} else {
							b = sizlib.ParseBool(ps.ByNameDflt(colName, "false"))
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10064): Invalid boolean data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false, ""
							}
							bp = AddBindValue(data, b)
							s = s + com + " " + hdlr.BindPlaceholder(bp)
						}
						n_gen++

					case "f": // Float
						var f float64
						nullOk := ValidateNullOk(v.ColName, ii, h, "insert")
						isNull := false
						if !ps.HasName(v.ColName) {
							isNull = true
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` NULL `)
						} else {
							f, err = strconv.ParseFloat(ps.ByNameDflt(colName, "0"), 64)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10065): Invalid Number column %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false, ""
							}
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10066): Invalid float data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false, ""
							}
							bp = AddBindValue(data, f)
							s = s + com + " " + hdlr.BindPlaceholder(bp)
						}
						n_gen++

					case "d": // Date
						fallthrough
					case "t": // Time
						fallthrough
					case "e": // Date-Time
						var d time.Time
						nullOk := ValidateNullOk(v.ColName, ii, h, "insert")
						isNull := false
						d, isNull, err = ms.FuzzyDateTimeParse(val, nullOk)
						if err != nil {
							hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10067): Invalid Time/Date column %s, %s, Value[%s]", err, v.ColName, ps.ByName(v.ColName)), err, v.ColName, trx, res, req, ps)
							return false, ""
						}
						if nullOk && isNull {
							s = s + com + fmt.Sprintf(` NULL `)
						} else {
							err = ValidateSetValue(v.ColName, ii, h, val, req)
							if err != nil {
								hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10068): Invalid Time/Date data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
								return false, ""
							}
							bp = AddBindValue(data, d)
							s = s + com + " " + hdlr.BindPlaceholder(bp)
						}
						n_gen++

					}
					com = ","
				} else if v.AutoGen {
					id, _ := uuid.NewV4()
					s_id = id.String()
					bp = AddBindValue(data, s_id)
					// s = s + com + " " + hdlr.BindPlaceholder(bp)
					if hdlr.GetDbType() == DbType_postgres {
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
					} else if hdlr.GetDbType() == DbType_odbc {
						s = s + com + ` convert(UniqueIdentifier,` + hdlr.BindPlaceholder(bp) + `)`
					} else if hdlr.GetDbType() == DbType_Oracle {
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
					} else {
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
					}
					err = ValidateSetValue(v.ColName, ii, h, s_id, req)
					if err != nil {
						hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10069): Invalid string data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
						return false, ""
					}
					com = ","
					n_gen++
				} else if v.DefaultData != "" {
					bp = AddBindValue(data, v.DefaultData)
					// s = s + com + " " + hdlr.BindPlaceholder(bp)
					if hdlr.GetDbType() == DbType_postgres {
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
					} else if hdlr.GetDbType() == DbType_odbc {
						s = s + com + ` convert(UniqueIdentifier,` + hdlr.BindPlaceholder(bp) + `)`
					} else if hdlr.GetDbType() == DbType_Oracle {
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
					} else {
						s = s + com + ` ` + hdlr.BindPlaceholder(bp)
					}
					err = ValidateSetValue(v.ColName, ii, h, v.DefaultData, req)
					if err != nil {
						hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10069): Invalid string data %s, %s, %v", err, v.ColName, h.LineNo), err, v.ColName, trx, res, req, ps)
						return false, ""
					}
					com = ","
					n_gen++
				}
			}
		}
		used["name"] = true // table name from URL
		for _, vv := range ps.Data[0:ps.NParam] {
			// xyzzyImprove - improve error message!
			if !used[vv.Name] && (vv.From == goftlmux.FromParams || vv.From == goftlmux.FromBody || vv.From == goftlmux.FromBodyJson) {
				fmt.Fprintf(os.Stderr, "%sNot Used: %s%s\n", MiscLib.ColorYellow, vv.Name, MiscLib.ColorReset)
				fmt.Printf("Warning: Supplied input but not used in insert statement: %s\n", vv.Name)
				trx.AddNote(2, fmt.Sprintf("Warning(19031): supplied in input to insert, but not used: %s", vv.Name))
			}
		}
		if len(h.CustomerIdPart.ColName) > 0 {
			bp := AddBindValue(data, ps.ByNameDflt("$customer_id$", "1"))
			s = s + com + " " + hdlr.BindPlaceholder(bp)
			com = ","
		}
		mdata["vals"] = s
		if sizlib.InArray("db_insert", h.DebugFlag) {
			fmt.Printf("Vals to insert are: %s, %s\n", s, godebug.LF())
		}
	} else {
		hdlr.CrudErrMsg(1, fmt.Sprintf("Error(18001): Configuration failed to specify insertable columns, LineNo:%v", h.LineNo), nil, "", trx, res, req, ps)
		return false, ""
	}

	if n_gen == 0 {
		hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10070): Insert must insert atleast 1 column %s", h.LineNo), nil, "", trx, res, req, ps)
		return false, ""
	} else if n_gen != n_col {
		hdlr.CrudErrMsg(1, fmt.Sprintf("Error(18070): Inserted columns(%d) did not match expected(%d) column %s", n_gen, n_col, h.LineNo), nil, "", trx, res, req, ps)
		return false, ""
	}

	ok = true
	return
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func CommonMakeData(h SQLOne, trx *tr.Trx, ps *goftlmux.Params) (data []interface{}) {
	data = make([]interface{}, len(h.P)) // organize data for call to d.b.
	for i, kk := range h.P {             // Not real clear to me why you would ever cobine CRUD and p:[] values
		data[i] = ps.ByName(kk)
		if !ps.HasName(kk) {
			trx.AddNote(2, fmt.Sprintf("Warning(10031): Missing data for %s, Empty string used!", kk))
		}
	}
	return
}

const db_common_crud1 = false

/* vim: set noai ts=4 sw=4: */
