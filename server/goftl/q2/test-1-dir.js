System.register(['angular2/core', './app.component'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
        var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
        if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
        else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
        return c > 3 && r && Object.defineProperty(target, key, r), r;
    };
    var __metadata = (this && this.__metadata) || function (k, v) {
        if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
    };
    var __param = (this && this.__param) || function (paramIndex, decorator) {
        return function (target, key) { decorator(target, key, paramIndex); }
    };
    var core_1, app_component_1;
    var Test1Dir;
    return {
        setters:[
            function (core_1_1) {
                core_1 = core_1_1;
            },
            function (app_component_1_1) {
                app_component_1 = app_component_1_1;
            }],
        execute: function() {
            Test1Dir = (function () {
                function Test1Dir(app) {
                    this.socket = null;
                    this.data2 = null;
                    this.model = {
                        TrxId: "",
                        SeqNo: ""
                    };
                    this.setIt = null;
                    this.user = "Test-User";
                    this.color = "red";
                    this.socket = app.getSio();
                    this.data2 = app.getData2();
                    this.setIt = app.setData2();
                }
                //sendMsg(msg:string){
                //	this.socket.emit('msg', JSON.stringify( { "name":"echo", "body":msg, "user":this.user } ));
                //}
                Test1Dir.prototype.FollowMostRecent = function () {
                    console.log("this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo);
                    this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/listen-for", "ClientTrxId": this.model.TrxId, "FilterId": "MostRecent", "maxKey": this.model.SeqNo }));
                };
                Test1Dir.prototype.SpecificId = function () {
                    console.log("this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo);
                    this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/listen-for", "ClientTrxId": this.model.TrxId, "FilterId": this.model.SeqNo, "maxKey": this.model.SeqNo }));
                };
                Test1Dir.prototype.FirstId = function () {
                    console.log("this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo);
                    this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/listen-for", "ClientTrxId": this.model.TrxId, "FilterId": "First", "maxKey": this.model.SeqNo }));
                };
                Test1Dir.prototype.LastId = function () {
                    console.log("this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo);
                    this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/listen-for", "ClientTrxId": this.model.TrxId, "FilterId": "Last", "maxKey": this.model.SeqNo }));
                };
                Test1Dir.prototype.DoPing = function () {
                    console.log("this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo);
                    this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/ping" }));
                };
                Test1Dir.prototype.DoDump = function () {
                    console.log("this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo);
                    this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/dump" }));
                };
                Test1Dir.prototype.DoAddData = function () {
                    this.setIt({
                        ResponseBytes: "",
                        Func: [
                            { FuncName: "aaafunc" },
                            { FuncName: "bbbfunc" },
                            { FuncName: "cccfunc" }
                        ],
                        Data: [],
                    });
                    this.data2 = {
                        ResponseBytes: "",
                        Func: [
                            { FuncName: "aaa_func" },
                            { FuncName: "bbb_func" },
                            { FuncName: "ccc_func" }
                        ],
                        Data: [],
                    };
                    ;
                };
                Test1Dir = __decorate([
                    core_1.Component({
                        selector: "test-1-dir",
                        template: "<div>\n<h3>Test for Tracer2 </h3> <br>\nTrxId: <input type=\"text\"  class=\"form-control\" required [(ngModel)]=\"model.TrxId\" > <br>\nSeqNo: <input type=\"text\"  class=\"form-control\" required [(ngModel)]=\"model.SeqNo\" > <br>\n<button (click)=\"FollowMostRecent()\"> FollowMostRecent </button>\n<button (click)=\"SpecificId()\"> SpecificId </button>\n<button (click)=\"FirstId()\"> FirstId </button>\n<button (click)=\"LastId()\"> LastId </button>\n<button (click)=\"DoPing()\"> Ping </button>\n<button (click)=\"DoDump()\"> Dump </button>\n<button (click)=\"DoAddData()\"> Do Da Ting </button>\n</div>\n\t\t\t<div class=\"dataTag\"><span> Functions Called: </span>\n\t\t\t\t<table class=\"dataOut\">\n\t\t\t\t\t<tr>\n\t\t\t\t\t\t<th>Name!</th>\n\t\t\t\t\t\t<th>File!</th>\n\t\t\t\t\t\t<th>Line!</th>\n\t\t\t\t\t</tr>\n\t\t\t\t\t<tr *ngFor=\"let x of data2.Func\">\n\t\t\t\t\t\t<td>{{x.FuncName}}</td>\n\t\t\t\t\t</tr>\n\t\t\t\t</table>\n\t\t\t</div>\n\n"
                    }),
                    __param(0, core_1.Inject(core_1.forwardRef(function () { return app_component_1.AppComponent; }))), 
                    __metadata('design:paramtypes', [app_component_1.AppComponent])
                ], Test1Dir);
                return Test1Dir;
            }());
            exports_1("Test1Dir", Test1Dir);
        }
    }
});
//# sourceMappingURL=test-1-dir.js.map