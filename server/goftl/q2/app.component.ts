
import {Component} from 'angular2/core';
import {Test1Dir} from "./test-1-dir"
import {TimerWrapper} from 'angular2/src/facade/async';

@Component({
	selector: 'sio-app',
	directives: [ Test1Dir ],
	template: `
<h1 class="main-page">Main Page H1</h1>
<test-1-dir></test-1-dir>
<hr>
<pre style="width:99%;max-width:99%;">
{{data | json}}
</pre>
<pre style="width:99%;max-width:99%;">
{{data2 | json}}
</pre>
			<div class="dataTag"><span> Functions Called: </span>
				<table class="dataOut">
					<tr>
						<th>Name</th>
						<th>File</th>
						<th>Line</th>
					</tr>
					<tr *ngFor="let x of data2.Func">
						<td>{{x.FuncName}}</td>
					</tr>
				</table>
			</div>
`
})

export class AppComponent {
	socket = null;
	data = null;
	data2 = {
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
 
	constructor() {
		this.socket = io(); // this.socket = io('http://localhost:16010');
		this.socket.on('/trx', function(data){
			// this.price = data;
			console.log ( "data =", data );
			this.data = JSON.parse(data);
			if (this.data && this.data.body) {
				this.data2 = JSON.parse(this.data.body);
				TimerWrapper.setTimeout(() => {  
					console.log('Send message that I finished painting');
					// } else if match, plist := DispatchMatch(mm, "rps", "/ready-for-more-data", "ClientTrxId"); match {
					this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/ready-for-more-data" } ));
				}, 250);
			}
		}.bind(this));	
	}

	getSio() {
		return this.socket;
	}

	getData2() {
		return this.data2;
	}

	setData2(d) {
		return function(d){
			console.log ( "Setting this.data2" );
			this.data2 = d;
		}
	}
 
	bShowY(x) {
		console.log ( "bShowY", x );
	}
}


