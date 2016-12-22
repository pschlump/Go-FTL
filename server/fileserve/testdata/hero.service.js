System.register(['angular2/core', 'angular2/http', 'rxjs/Observable', 'rxjs/Rx'], function(exports_1, context_1) {
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
    var core_1, http_1, http_2, Observable_1;
    var HeroService;
    return {
        setters:[
            function (core_1_1) {
                core_1 = core_1_1;
            },
            function (http_1_1) {
                http_1 = http_1_1;
                http_2 = http_1_1;
            },
            function (Observable_1_1) {
                Observable_1 = Observable_1_1;
            },
            function (_1) {}],
        execute: function() {
            HeroService = (function () {
                function HeroService(http) {
                    this.http = http;
                    /*
                    // private _heroesUrl = 'app/heroes.json'; // URL to JSON file
                    // private _heroesUrl = 'app/heroes';	// URL to web api
                    // private _heroesUrl = 'data.hero.json';	// URL to web api -- as a static file
                    */
                    this._heroesUrl = '/api/table/heros'; // URL to web api
                }
                HeroService.prototype.getHeroes = function () {
                    return this.http.get(this._heroesUrl)
                        .map(this.extractData)
                        .catch(this.handleError);
                };
                HeroService.prototype.addHero = function (name) {
                    var body = JSON.stringify({ name: name });
                    var headers = new http_2.Headers({ 'Content-Type': 'application/json' });
                    var options = new http_2.RequestOptions({ headers: headers });
                    return this.http.post(this._heroesUrl, body, options)
                        .map(this.extractData)
                        .catch(this.handleError);
                };
                // delHero - xyzzy
                HeroService.prototype.extractData = function (res) {
                    if (res.status < 200 || res.status >= 300) {
                        throw new Error('Bad response status: ' + res.status);
                    }
                    // xyzzy - check "status":"success" - if not error
                    // xyzzy - prefix check
                    // xyzzy - handle encrypted data at this point if ( "X-Encrypted" header? )
                    var body = res.json();
                    // console.log ( "body=", body )
                    // xyzzy - if "Meta" configured then use {} , else []
                    // this.heroes = body.data || { };			// meta data format migth be { "data": [ ... ], "nRows": 10 }
                    // return body.data || { };
                    this.heroes = body.data || [];
                    // console.log ( "Saved data in this.heroes:", this.heroes );
                    return body.data || [];
                };
                HeroService.prototype.updHero = function (hero) {
                    var body = JSON.stringify({ hero: hero });
                    var headers = new http_2.Headers({ 'Content-Type': 'application/json' });
                    var options = new http_2.RequestOptions({ headers: headers });
                    console.log("updHero - service", body);
                    var cacheBlow = "?_=" + Math.random();
                    var url = this._heroesUrl + cacheBlow + "&METHOD=PUT";
                    console.log("updHero - url", url);
                    return this.http.put(url, body, options)
                        .map(this.extractData)
                        .catch(this.handleError);
                };
                /*
                    http://blog.thoughtram.io/angular/2016/03/21/template-driven-forms-in-angular-2.html
                    http://chariotsolutions.com/blog/post/angular2-observables-http-separating-services-components/
                */
                /*
                    // getHero(id: number) {
                    // getHero(id: string) {
                    getHero(id: string): Observable<Hero[]> {
                        //return Promise.resolve(HEROES).then(
                        //	heroes => heroes.filter(hero => hero.id === id)[0]
                        //);
                    console.log ( 'getHero' );
                        //return Promise.resolve(this.heroes).then(
                        //	heroes => {
                        //		// heroes.filter(hero => hero.id === id)[0]
                        //		for ( var i = 0; i < heroes.length; i++ ) {
                        //			if ( heroes[i].id === id ) {
                        //				return heroes[i];
                        //			}
                        //		}
                        //	}
                        //);
                        return this.getHeroes()
                            .subscribe(heroes => {
                                    // this.heroes = heroes.slice(1,5);
                    console.log ( 'getHero in callback', heroes );
                                    for ( var i = 0; i < heroes.length; i++ ) {
                                        if ( heroes[i].id === id ) {
                                            return heroes[i];
                                        }
                                    }
                                }
                                // , error => this.errorMessage = <any>error
                            );
                    }
                */
                HeroService.prototype.handleError = function (error) {
                    // In a real world app, we might send the error to remote logging infrastructure -- /api/table/send-logger
                    var errMsg = error.message || 'Server error';
                    console.error(errMsg); // log to console instead
                    return Observable_1.Observable.throw(errMsg);
                };
                HeroService = __decorate([
                    core_1.Injectable(), 
                    __metadata('design:paramtypes', [http_1.Http])
                ], HeroService);
                return HeroService;
            }());
            exports_1("HeroService", HeroService);
        }
    }
});
//# sourceMappingURL=hero.service.js.map