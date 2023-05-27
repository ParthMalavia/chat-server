import { Component, OnInit, OnDestroy } from '@angular/core';
import { SocketService } from './socket.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit, OnDestroy {
  title = 'ang-chat-client';

  public messages: Array<any>;
  public chatBox: string;

  public constructor(private socket: SocketService) {
    this.messages = [];
    this.chatBox = "";
  }

  public ngOnInit(): void {
    this.socket.getEventListener().subscribe(event => {
      // switch (event.type) {
      //   case "message": {
      //     let data = event.data.Content;
      //     if (event.data.sender) {
      //       data = event.data.sender + ": " + data;
      //     }
      //     console.log(data);
      //     this.messages.push(data);
      //     break;
      //   }
      //   case "open": {
      //     this.messages.push("/The socket connection has been established.");
      //     break;
      //   }
      //   case "close": {
      //     this.messages.push("/The socket connection has been close.");
      //     break;
      //   }
      // }
      if (event.type==="message") {
        let data = event.data.content;
          if (event.data.sender) {
            data = event.data.name + ": " + data;
          }
          console.log(data);
          this.messages.push(data);
      }

      if (event.type==="open") {
        this.messages.push("/The socket connection has been established.");
      }

      if (event.type==="close") {
        this.messages.push("/The socket connection has been close.");
      }
    })
  }

  public ngOnDestroy(): void {
      this.socket.close()
  }

  public send() {
    if (this.chatBox) {
      this.socket.send(this.chatBox);
      this.chatBox = "";
    }
  }

  public isSystemMessage(message: string) {
    return message.startsWith("/") ? "<strong>" +message.substring(1)+ "</strong>" : message;
  }
}
