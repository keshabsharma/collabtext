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
      room: "1",
      client: this.makeid(10),
      onLine: navigator.onLine,
      showBackOnline: true,
      revision: 0
    }
  },

  mounted() {
    // Start up editor
    console.log('this is current quill instance object', this.editor);

    // Start WebSocket Connection
    this.createConnection()

    // Make Call to set information
    // this.getFiles()

    window.addEventListener('online', this.updateOnlineStatus);
    window.addEventListener('offline', this.updateOnlineStatus);


    // this.timer = setInterval(() => {
    //   if (this.showBackOnline) {
    //     var ot = this.log.shift()
    //     if (ot != undefined) {
    //       if ( (this.connection.readyState === WebSocket.CLOSED) || !(this.showBackOnline) ) {
    //         console.log("can't push right now")
    //       } else {
    //         console.log("pushed from log", ot)
    //       }
    //     }
    //   } else {
    //     console.log("can't push right now")
    //   }
    // }, 500)


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
        this.$refs.myQuillEditor.quill.once('text-change', (delta) =>  {
          console.log("Delta change!: ", delta)
          var ops = delta['ops']
          var mostrecent = ops[1]
          var val = { 
            "revision": this.revision, 
            "op": "",
            "position": 0,
            "str": "",
            "client": this.client,
            "document": "",
            "error": "" 
          }
          if ('delete' in mostrecent) {
            val = { 
              "revision": this.revision, 
              "op": "delete",
              "position": (ops[0]['retain'] + mostrecent["delete"]),
              "str": mostrecent["delete"],
              "client": this.client,
              "document": this.room,
              "error": ""
            }
            // console.log(val)
          } else if ('insert' in mostrecent) {
            val = { 
              "revision": this.revision, 
              "op": "insert",
              "position": ops[0]['retain'],
              "str": mostrecent["insert"],
              "client": this.client,
              "document": this.room,
              "error": ""
            }
            // console.log(val)
          }
          console.log(val)
          this.log.push(val)
          this.revision += 1
          console.log(this.revision)

          // if ('attributes' in mostrecent) {
          //   val = "attribute: " + mostrecent["attributes"] + " @ " + ops[0]['retain']            
          //   console.log(val)
          // }

        });
        this.content = html
      },

      receiveMessage (message) {
        if (message.client != this.client) {
          var protocol = message.op
          var str = message.str
          var index = message.position
          // var protocol = message.substr(0, message.indexOf(":"))
          // var str= message.substr(message.indexOf(':')+1, message.lastIndexOf('@') - message.indexOf(':') - 1);
          // var index = message.substr(message.lastIndexOf('@') + 1)
          switch(protocol) {
            // case "attr":
            //   // this.$refs.myQuillEditor.quill.formatText(0, 5, 'bold', false);
            //   break;
            case "delete":
              this.$refs.myQuillEditor.quill.deleteText(index+str, str);
              break;
            case "insert":
              this.$refs.myQuillEditor.quill.insertText(index, str, 'bold', false);
              break;
            default:
              
          }
          if (this.revision < message.revision) {
            this.revision = message.revision
          }

        } else {
          var ot = this.log.shift()
          if (ot !== undefined) {
            if (this.connection.readyState === WebSocket.CLOSED) {
              console.log("can't push right now")
            } else {
              this.connection.send(ot)
              console.log("pushed from log", ot)
            }
          }
        }
      },


      resetDocument(data) {
        console.log(data)
        this.$refs.myQuillEditor.quill.setContents([{ insert: 'hi\n' }]);
      },


      createConnection () {
        console.log("Starting connection to WebSocket Server")
        this.connection = new WebSocket("ws://localhost:7777/ws/" + this.room)

        this.connection.onopen = function(event) {
          console.log(event)
          console.log("Successfully connected to the echo websocket server...")
        }

        this.connection.onmessage = function(event) {
          console.log(event.data);
          this.receiveMessage(event.data);
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
        // this.onLine = navigator.onLine;
        if (this.onLine !== this.showBackOnline) {
          if (this.onLine) {
            this.createConnection()
            if (this.connection.readyState !== WebSocket.CLOSED) {
              this.resetDocument("test")
              this.showBackOnline = true
            }
          } else {
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
      }

      // getFiles () {
        // axios
        //   .get('URL')
        //   .then(response => (this.deck = response))
      // }

      // getData() {
        // axios
        //   .get('URL')
        //   .then(response => (this.deck = response))
      // }

      // dosomething () {
      //   this.$refs.myQuillEditor.quill.insertText(0, 'a', 'bold', true);
      // },

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
