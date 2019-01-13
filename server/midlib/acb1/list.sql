--
--
-- 	,"/api/pullDataTopUser": { "query": "select * from \"e_note\" as t1 where exists ( select 'found' as \"found\" from \"e_page\" as t2 where t2.\"id\" = t1.\"page_id\" and t2.\"url\" = $1 ) ", "p": [ "url" ]
-- 		, "LineNo":"29"
-- 		, "valid": {
-- 			 "url": { "required":true, "type":"string" }
-- 			,"user": { "required":true, "type":"string", "min_len":4, "max_len":100 }
-- 			,"auth_token": { "required":true, "type":"uuid" }
-- 			,"callback": { "optional":true }
-- 			}
-- 		}
--
--		, "SelectPK1Tmpl": " SELECT \"tblActionPlan\".\"id\" ,\"tblActionPlan\".\"cardId\" ,\"tblActionPlan\".\"sequence\" ,\"tblActionPlan\".\"actionPlan\" ,\"tblActionPlan\".\"dateEntered\" ,\"tblActionPlan\".\"targetCompletion\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"notes\" ,\"tblActionPlan\".\"actionCompleted\" ,\"tblActionPlan\".\"isDeleted\" ,\"tblPerson\".\"firstName\" ,\"tblPerson\".\"lastName\" ,\"tblPerson\".\"email\" ,\"tblPerson\".\"phone\" FROM \"tblActionPlan\" as \"tblActionPlan\" left join \"tblPerson\" as \"tblPerson\" on \"tblActionPlan\".\"responsiblePersonId\" = \"tblPerson\".\"id\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"




select t1.*
	, t2."file_name"
	, t2."url_path"
	, t2."qr_id"
	, t2."qr_enc_id"
	, t2."state" as "qr_state"
from "v1_trackAdd" as t1 left outer join "v1_avail_qr" as t2 on t1."qr_id" = t2."id"
;
