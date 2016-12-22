import {Injectable}		         from 'angular2/core';
import {Http, Response}          from 'angular2/http';
import {Headers, RequestOptions} from 'angular2/http';
import {Hero}					 from './hero';
import {Observable}		         from 'rxjs/Observable';
import 'rxjs/Rx';

@Injectable()
export class HeroService {
	constructor (private http: Http) {}

	/*
	// private _heroesUrl = 'app/heroes.json'; // URL to JSON file
	// private _heroesUrl = 'app/heroes';	// URL to web api
	// private _heroesUrl = 'data.hero.json';	// URL to web api -- as a static file
	*/

	private _heroesUrl = '/api/table/heros';	// URL to web api
	// private heroes;
	private heroes: Hero[];

	getHeroes (): Observable<Hero[]> {
		return this.http.get(this._heroesUrl)
					.map(this.extractData)
					.catch(this.handleError);
	}

	addHero (name: string): Observable<Hero>	{

		let body = JSON.stringify({ name });
		let headers = new Headers({ 'Content-Type': 'application/json' });
		let options = new RequestOptions({ headers: headers });

		return this.http.post(this._heroesUrl, body, options)
					.map(this.extractData)
					.catch(this.handleError);
	}

	// delHero - xyzzy

	private extractData(res: Response) {
		if (res.status < 200 || res.status >= 300) {
			throw new Error('Bad response status: ' + res.status);
		}
		// xyzzy - check "status":"success" - if not error
		// xyzzy - prefix check
		// xyzzy - handle encrypted data at this point if ( "X-Encrypted" header? )
		let body = res.json();
		// console.log ( "body=", body )
		// xyzzy - if "Meta" configured then use {} , else []
		// this.heroes = body.data || { };			// meta data format migth be { "data": [ ... ], "nRows": 10 }
		// return body.data || { };
		this.heroes = body.data || [];
		// console.log ( "Saved data in this.heroes:", this.heroes );
		return body.data || [];
	}

	updHero( hero : Hero ) {

		let body = JSON.stringify({ hero });
		let headers = new Headers({ 'Content-Type': 'application/json' });
		let options = new RequestOptions({ headers: headers });

		console.log ( "updHero - service", body );

		let cacheBlow = "?_=" + Math.random();
		let url = this._heroesUrl + cacheBlow + "&METHOD=PUT";

		console.log ( "updHero - url", url );

		return this.http.put(url, body, options)
					.map(this.extractData)
					.catch(this.handleError);
	}
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

	private handleError (error: any) {
		// In a real world app, we might send the error to remote logging infrastructure -- /api/table/send-logger
		let errMsg = error.message || 'Server error';
		console.error(errMsg); // log to console instead
		return Observable.throw(errMsg);
	}
}

