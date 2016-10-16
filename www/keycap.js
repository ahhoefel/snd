
var socket;

var debug = false;

var command = {
  IMAGE: 0,
  TEXT: 1,
  MIDI: 2,
  KNOB_BANK: 3,
  KNOB_VALUES: 4
};

var knobBanks = [ 
{ "name": "Waveform",
  "labels": ["Spike", "Skew", "Flat", "Empty", "Overtones", "Spacing", "Decay", "Empty"],
},
{ "name": "Delay",
  "labels": ["Delays", "DelayTime", "DelayDecay", "Empty", "Empty", "Empty", "Empty", "Empty"],
}
];

var knobBank = knobBanks[0];

function init() {
  var charfield = document.getElementById("char");
  var out = document.getElementById("out");
  createKnobForm();

  socket = new WebSocket("ws://localhost:8080/echo", ["arraybuffer"]);
  socket.onmessage = onMessage;
  // charfield.onkeypress = function(e) {
   //  e = e || window.event;
    // var charCode = (typeof e.which == "number") ? e.which : e.keyCode;
     //if (charCode > 0) {
     //  out.innerHTML = String.fromCharCode(charCode);
      // socket.send(String.fromCharCode(charCode));
    // }
   //};
  initMidi();
}
window.onload = init;

function createKnobForm() {
  console.log("Creating knob form");
  setKnobNext();
  var knobForm = document.getElementById("knobForm");
  for (var i = 0; i < knobBanks.length; i++) {
    p = knobBanks[i];
    console.log("Bank "+ p);
    var fs = document.createElement("fieldset");
    p.legendText = document.createTextNode(p.name);
    p.legend = document.createElement("legend");
    p.legend.appendChild(p.legendText);
    p.legend.id = "knobBankLegend" + p.title;
    p.knobs = [];
    fs.appendChild(p.legend);
    for (r = 0; r < 2; r++) {
      var row = document.createElement("div");
      for (j = 0; j < 4; j++) {
        var k = document.createElement("span");
        k.id = "knob" + (4*r + j);
        k.className = "knob";
        k.label = p[4*r+j];
        row.appendChild(k);
        p.knobs.push(k); 
      }
      fs.appendChild(row);
    }
    for (var k = 0; k < p.knobs.length; k++) {
        setKnobValue(p, k, 0);
    }
    knobForm.appendChild(fs);
  }
}

function setKnobNext() {
  for (var i = 0; i < knobBanks.length; i++) {
    knobBanks[i].next = knobBanks[(i+1)%knobBanks.length];
  }
}

function setKnobValues(values) {
  for (var i = 0; i < values.length; i++) {
    var k = values[i];
    for (var j = 0; j < k.length; j++) {
      setKnobValue(knobBanks[i], j, k[j]);
    }
  }
}

Number.prototype.toFixedDown = function(digits) {
    var re = new RegExp("(\\d+\\.\\d{" + digits + "})(\\d)"),
        m = this.toString().match(re);
    return m ? parseFloat(m[1]) : this.valueOf();
};

function setKnobValue(bank, k, v) {
  (bank.knobs)[k].innerHTML = bank.labels[k] + ": " + v.toFixedDown(3);
}


function nextKnobBank() {
  knobBank.legend.innerHTML = knobBank.name;
  knobBank = knobBank.next;
  knobBank.legend.innerHTML = "<b>" + knobBank.name + "</b>";
}


function onMessage(event) {
  var m = JSON.parse(event.data);
  //console.log(m)
  if (m.Cmd == command.TEXT) {
    $('div#output').html(window.atob(m.Data));
  } else if (m.Cmd == command.IMAGE) {
    drawImage(window.atob(m.Data));
  } else if (m.Cmd == command.KNOB_BANK) {
    nextKnobBank();
  } else if (m.Cmd == command.KNOB_VALUES) {
    console.log(window.atob(m.Data));
    setKnobValues(JSON.parse(window.atob(m.Data)));
  }
}

function drawImage(data) {
  var c = document.getElementById("outputimage");
  var ctx = c.getContext("2d");
  var imgData=ctx.createImageData(200,100);
  for (var i=0; i < imgData.data.length; i++) {
    imgData.data[i] = data.charCodeAt(i);
  }
  ctx.putImageData(imgData,0,0);
}


function initMidi() {
  // request MIDI access
  if (navigator.requestMIDIAccess) {
    navigator.requestMIDIAccess({
      sysex: false // this defaults to 'false' and we won't be covering sysex in this article. 
    }).then(onMIDISuccess, onMIDIFailure);
  } else {
    alert("No MIDI support in your browser.");
  }
}

// midi functions
function onMIDISuccess(midiAccess) {
    // when we get a succesful response, run this code
    console.log('MIDI Access Object', midiAccess);
    // when we get a succesful response, run this code
    midi = midiAccess; // this is our raw MIDI data, inputs, outputs, and sysex status

    var inputs = midi.inputs.values();
    // loop over all available inputs and listen for any MIDI input
    for (var input = inputs.next(); input && !input.done; input = inputs.next()) {
        // each time there is a midi message call the onMIDIMessage function
        input.value.onmidimessage = onMIDIMessage;
    }
}

function onMIDIFailure(e) {
    // when we get a failed response, run this code
    console.log("No access to MIDI devices or your browser doesn't support WebMIDI API. Please use WebMIDIAPIShim " + e);
}

function onMIDIMessage(message) {
  data = message.data; // this gives us our [command/channel, note, velocity] data.
  if (data[0] == 192 && data[2] == 0) {
    debug = !debug;
    console.log('Debug:', debug); 
  }
  if (debug) {
    console.log('MIDI data', data); // MIDI data [144, 63, 73]
  }
  //console.log('Sending', String.fromCodePoint(data[0],data[1],data[2]))
  //socket.send(String.fromCodePoint(data[0],data[1],data[2]))
  var buf = new ArrayBuffer(3)
  var bufView = new Uint8Array(buf)
  bufView[0] = data[0]
  bufView[1] = data[1]
  bufView[2] = data[2]
  //console.log('Sending', bufView)
  socket.send(buf)
}

