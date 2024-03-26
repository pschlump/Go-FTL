System.register(['angular2/core', "./test-1-dir", 'angular2/src/facade/async'], function(exports_1, context_1) {
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
    var core_1, test_1_dir_1, async_1;
    var AppComponent;
    return {
        setters:[
            function (core_1_1) {
                core_1 = core_1_1;
            },
            function (test_1_dir_1_1) {
                test_1_dir_1 = test_1_dir_1_1;
            },
            function (async_1_1) {
                async_1 = async_1_1;
            }],
        execute: function() {
            AppComponent = (function () {
                function AppComponent() {
                    this.socket = null;
                    this.data = null;
                    this.data2 = {
                        Uri: "",
                        From: "",
                        Status: "",
                        Method: "",
                        ClientIp: "",
                        RequestTime: "",
                        ElapsedTimeMs: "",
                        RvBody: "",
                        ResponseBytes: "",
                        Func: [],
                        Data: [],
                        Qry: [],
                        TableList: [],
                        Note: []
                    };
                    this.socket = io(); // this.socket = io('http://localhost:16010');
                    this.socket.on('/trx', function (data) {
                        var _this = this;
                        // this.price = data;
                        console.log("data =", data);
                        this.data = JSON.parse(data);
                        if (this.data && this.data.body) {
                            this.data2 = JSON.parse(this.data.body);
                            async_1.TimerWrapper.setTimeout(function () {
                                console.log('Send message that I finished painting');
                                // } else if match, plist := DispatchMatch(mm, "rps", "/ready-for-more-data", "ClientTrxId"); match {
                                _this.socket.emit('msg', JSON.stringify({ "To": "rps://tracer/ready-for-more-data" }));
                            }, 250);
                        }
                    }.bind(this));
                }
                AppComponent.prototype.getSio = function () {
                    return this.socket;
                };
                AppComponent.prototype.getData2 = function () {
                    return this.data2;
                };
                AppComponent.prototype.setData2 = function (d) {
                    return function (d) {
                        console.log("Setting this.data2");
                        this.data2 = d;
                    };
                };
                AppComponent.prototype.bShowY = function (x) {
                    console.log("bShowY", x);
                };
                AppComponent = __decorate([
                    core_1.Component({
                        selector: 'sio-app',
                        directives: [test_1_dir_1.Test1Dir],
                        template: "\n<h1 class=\"main-page\">Main Page H1</h1>\n<test-1-dir></test-1-dir>\n<hr>\n<pre style=\"width:99%;max-width:99%;\">\n{{data | json}}\n</pre>\n<pre style=\"width:99%;max-width:99%;\">\n{{data2 | json}}\n</pre>\n\t\t\t<div class=\"dataTag\"><span> Functions Called: </span>\n\t\t\t\t<table class=\"dataOut\">\n\t\t\t\t\t<tr>\n\t\t\t\t\t\t<th>Name</th>\n\t\t\t\t\t\t<th>File</th>\n\t\t\t\t\t\t<th>Line</th>\n\t\t\t\t\t</tr>\n\t\t\t\t\t<tr *ngFor=\"let x of data2.Func\">\n\t\t\t\t\t\t<td>{{x.FuncName}}</td>\n\t\t\t\t\t</tr>\n\t\t\t\t</table>\n\t\t\t</div>\n"
                    }), 
                    __metadata('design:paramtypes', [])
                ], AppComponent);
                return AppComponent;
            }());
            exports_1("AppComponent", AppComponent);
        }
    }
});
//# sourceMappingURL=app.component.js.map