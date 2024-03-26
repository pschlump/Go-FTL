import {Component, Injectable, Inject, forwardRef} from 'angular2/core';
import {AppComponent} from './app.component';

@Component({
	selector: "test-1-dir",
	template: `<div>
<h3>Test for Tracer2 </h3> <br>
TrxId: <input type="text"  class="form-control" required [(ngModel)]="model.TrxId" > <br>
SeqNo: <input type="text"  class="form-control" required [(ngModel)]="model.SeqNo" > <br>
<button (click)="FollowMostRecent()"> FollowMostRecent </button>
<button (click)="SpecificId()"> SpecificId </button>
<button (click)="FirstId()"> FirstId </button>
<button (click)="LastId()"> LastId </button>
<button (click)="DoPing()"> Ping </button>
<button (click)="DoDump()"> Dump </button>
<button (click)="DoAddData()"> Do Da Ting </button>
</div>
			<div class="dataTag"><span> Functions Called: </span>
				<table class="dataOut">
					<tr>
						<th>Name!</th>
						<th>File!</th>
						<th>Line!</th>
					</tr>
					<tr *ngFor="let x of data2.Func">
						<td>{{x.FuncName}}</td>
					</tr>
				</table>
			</div>

`
})

export class Test1Dir {
	user:string;
	color:string;
	socket = null;
	data2 = null;
	model = {
		TrxId:"",
		SeqNo:""
	};
	setIt = null;
	constructor(@Inject(forwardRef(() => AppComponent)) app:AppComponent) {
		this.user = "Test-User";
		this.color = "red";
		this.socket = app.getSio();
		this.data2 = app.getData2();
		this.setIt = app.setData2();
	}
	//sendMsg(msg:string){
	//	this.socket.emit('msg', JSON.stringify( { "name":"echo", "body":msg, "user":this.user } ));
	//}
	FollowMostRecent(){
		console.log ( "this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo );
		this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/listen-for", "ClientTrxId":this.model.TrxId, "FilterId":"MostRecent", "maxKey":this.model.SeqNo } ));
	}
	SpecificId(){
		console.log ( "this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo );
		this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/listen-for", "ClientTrxId":this.model.TrxId, "FilterId":this.model.SeqNo, "maxKey":this.model.SeqNo } ));
	}
	FirstId(){
		console.log ( "this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo );
		this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/listen-for", "ClientTrxId":this.model.TrxId, "FilterId":"First", "maxKey":this.model.SeqNo } ));
	}
	LastId(){
		console.log ( "this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo );
		this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/listen-for", "ClientTrxId":this.model.TrxId, "FilterId":"Last", "maxKey":this.model.SeqNo } ));
	}
	DoPing(){
		console.log ( "this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo );
		this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/ping"}));
	}
	DoDump(){
		console.log ( "this.model.TrxId=", this.model.TrxId, " this.model.SeqNo=", this.model.SeqNo );
		this.socket.emit('msg', JSON.stringify( { "To":"rps://tracer/dump"}));
	}
	DoAddData() {
		this.setIt( {
				ResponseBytes: "",
				Func: [
					{ FuncName:"aaafunc" },
					{ FuncName:"bbbfunc" },
					{ FuncName:"cccfunc" }
				],
				Data: [],
			} );
		this.data2 = {
				ResponseBytes: "",
				Func: [
					{ FuncName:"aaa_func" },
					{ FuncName:"bbb_func" },
					{ FuncName:"ccc_func" }
				],
				Data: [],
			} );
	} 
}


