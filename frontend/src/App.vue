<template>
  <div id="app">
    <div class="column1">
      <li v-for="item in deck" :key="item" v-on:click="changeRoom(item)" class="textlink"><a>{{ item }}</a></li>
    </div>
    <div class="column2">
      <quill-editor
        ref="myQuillEditor"
        v-model="content"
        :options="editorOption"
        @blur="onEditorBlur($event)"
        @focus="onEditorFocus($event)"
        @ready="onEditorReady($event)"
        @change="onEditorChange($event)"
      />
    </div>
    <button @click="changeOnline">Start timer...</button>
  </div>

</template>

<script>
import 'quill/dist/quill.core.css'
import 'quill/dist/quill.snow.css'
import 'quill/dist/quill.bubble.css'
import { quillEditor } from 'vue-quill-editor'
import axios from 'axios'

export default {

  name: 'App',

  components: {
    quillEditor
  },

  data () {
    return {
      content: '<p>"I am here"</p>',
      editorOption: {
        
      },
      connection: null,
      log: [],
      deck: ["file1", "file2", "file3"],
      room: "doc1",
      client: this.makeid(10),
      onLine: navigator.onLine,
      showBackOnline: false,
      revision: 1,
      doc: ""
    }
  },

  mounted() {
    // Start up editor
    console.log('this is current quill instance object', this.editor);

    // Start WebSocket Connection
    this.updateOnlineStatus()
    // this.createConnection()
    // Make Call to set information
    // this.getFiles()

    window.addEventListener('online', this.updateOnlineStatus);
    window.addEventListener('offline', this.updateOnlineStatus);


  },


  watch: {
      onLine: function() {
        this.updateOnlineStatus();
      }
  },


  methods: {
      onEditorBlur(quill) {
        console.log('editor blur!', quill)
      },

      onEditorFocus() {
        var range = this.$refs.myQuillEditor.quill.getSelection();
        console.log('Cursor Index: ', range) // gets the cursor
      },

      onEditorReady(quill) {
        console.log('editor ready!', quill)
      },

      onEditorChange({ html }) {
        this.$refs.myQuillEditor.quill.once('text-change', (delta, oldDelta, source) =>  {
          if (source != 'api') {
            // var range = this.$refs.myQuillEditor.quill.getSelection();
            // console.log(source)
            // console.log("Delta change!: ", delta)
            // console.log("Current Cursor: ", range.index)
            var ops = delta['ops']
            var mostrecent = ops[1]
            if (ops.length == 1) {
              mostrecent = ops[0]
            }

            var retain = 0
            if (ops[0]["retain"]) {
              retain = ops[0]["retain"]
            }
            
            var val = { 
              "revision": this.revision, 
              "op": "",
              "position": 0,
              "str": "",
              "client": this.client,
              "document": this.room,
              "error": "" 
            }
            if ('delete' in mostrecent) {
              val = { 
                "revision": this.revision+1, 
                "op": "delete",
                "position": retain,
                "str": "",
                "client": this.client,
                "document": this.room,
                "error": ""
              }
            } else if ('insert' in mostrecent) {
              val = { 
                "revision": this.revision+1, 
                "op": "insert",
                "position": retain,
                "str": mostrecent["insert"],
                "client": this.client,
                "document": this.room,
                "error": ""
              }
            }

            //console.log(val)
            this.log.push(val)
            if (this.log.length == 1) {
                var ot = this.log.shift()
                if (ot !== undefined) {
                  if (this.connection.readyState === WebSocket.CLOSED) {
                    //console.log("can't push right now")
                  } else {
                    this.connection.send(JSON.stringify(ot))
                    //console.log("pushed from log", ot)
                  }
                }
            }
            //this.revision += 1
            //console.log(this.revision)
          }
        });
        this.content = html
      },

      receiveMessage (datas) {
        //console.log("showing message")
        var message = JSON.parse(datas)
        console.log(JSON.parse(datas))

        if (message.error) {
          console.log(message.error)
          return;

        }

        if (this.revision < message.revision) {
          this.revision = message.revision
        }
        if (message.client != this.client) {
          var protocol = message.op
          var str = message.str
          var index = message.position
          console.log("right here")
          console.log(this.$refs)
          switch(protocol) {
            case "delete":
              this.$refs.myQuillEditor.quill.deleteText(index, 1, 'api');
              break;
            case "insert":
              this.$refs.myQuillEditor.quill.insertText(index, str);
              break;
            default:
          }
        } else {
          this.shiftLog();
        }
      },

      shiftLog() {
        var ot = this.log.shift()
          if (ot !== undefined) {
            if (this.connection.readyState === WebSocket.CLOSED) {
              //console.log("can't push right now")
            } else {
              this.connection.send(JSON.stringify(ot))
              //console.log("pushed from log", ot)
            }
          }
      },

      resetDocument(data) {
        console.log(data)
        this.$refs.myQuillEditor.quill.setContents([{ insert: '\n' }]);
        // TODO: Asdd function to pull data from the server.
        this.getData(this.room)
        this.$refs.myQuillEditor.quill.setText(this.doc)
      },


      createConnection () {
        var _self = this;
        console.log("Starting connection to WebSocket Server")
        // TODO: Change connection url
        // this.connection = new WebSocket("ws://localhost:7778")
        this.connection = new WebSocket("ws://localhost:7777/ws/doc1")

        this.connection.onopen = function(event) {
          console.log(event)
          console.log("Successfully connected to the echo websocket server...")
        }

        this.connection.onmessage = function(event) {
          _self.receiveMessage(event.data);
        }

        this.connection.onclose = function(event) {
          console.log('Server is down.', event.reason);
        }
      },


      changeRoom(roomfile) {
        this.connection.close()
        this.room = roomfile
        this.createConnection()
      },


      updateOnlineStatus() {
        this.onLine = navigator.onLine;
        if (this.onLine !== this.showBackOnline) {
          if (this.onLine) {
            //console.log("User just went back online. Reconnecting to server")
            this.createConnection()
            if (this.connection.readyState !== WebSocket.CLOSED) {
              this.resetDocument("")
              this.showBackOnline = true
            }
          } else {
            //console.log("User just went offline. Stopping all syncronization")
            this.connection.close()
            this.showBackOnline = false
          }
        }
      },


      changeOnline() {
        this.onLine = !(this.onLine)
        console.log(this.onLine)
      },


      makeid(length) {
        var result           = [];
        var characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        var charactersLength = characters.length;
        for ( var i = 0; i < length; i++ ) {
            result.push(characters.charAt(Math.floor(Math.random() * charactersLength)));
        }
        return result.join('');
      },

      // getFiles () {
        // axios
        //   .get('URL')
        //   .then(response => (this.deck = response))
      // }

      getData(room) {
        axios
          .get('http://localhost:7777/document/' + room)
          .then(response => {
            this.revision = response.data.revision 
            this.$refs.myQuillEditor.quill.setText(response.data.document)
            this.doc = response.data.document
          })
          .catch(function (error) {
            // handle error
            console.log(error);
          })

      }

  },

  computed: {
    editor() {
      return this.$refs.myQuillEditor.quill
    }
  },

}
</script>


<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
.column1 {
  float: left;
  width: 15%;
}

/* Clear floats after the columns */
.column2 {
  float: right;
  width: 85%;
}

.textlink,
.textlink:active,
.textlink:focus,
.textlink:hover,
.textlink:visited {
    color: inherit;
    text-decoration: none;
    cursor: default;
}

</style>
