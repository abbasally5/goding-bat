// .js code for the challenge page
// Written in a simple/primitive way


var re = new XMLHttpRequest();
var lastMS = 0;

// Send code front-end
function sendCodeAce(){
  var now = new Date().getTime();
  // Mild double-click elimination
  if (lastMS!=0 && (now-lastMS) < 400) {
    return;
  }
  lastMS = now;
  sendCodeAce2();
}

// new ace-aware version
function sendCodeAce2(){
  var data = 'id='   + encodeURIComponent(document.codeform.id.value) + '&' +
         'code=' + encodeURIComponent(document.ace_editor.getValue()) + '&' +
         'cuname=' + encodeURIComponent(document.codeform.cuname.value) + '&' +
         'owner=' + encodeURIComponent(document.codeform.owner.value);
  if (document.codeform.parent) {
    data +=  '&parent=' + encodeURIComponent(document.codeform.parent.value);
  } 
  if (document.codeform.outputonly.checked) {
    data += '&outputonly=1';
  }
  if (document.codeform.date) {
    data +=  '&date=' + encodeURIComponent(document.codeform.date.value);
  }
  if (document.codeform.expnext) {  // exp
    data +=  '&expnext=' + encodeURIComponent(document.codeform.expnext.value);
  }
  if (document.codeform.adate) {  // exp
    data +=  '&adate=' + encodeURIComponent(document.codeform.adate.value);
  }
  if (document.ace_font) {  // 2017
    data +=  '&font=' + encodeURIComponent(document.ace_font);
  }
  if(re.readyState > 0 && re.readyState < 4) {  // do nothing if already in flight
    return true;
  }
  re.onreadystatechange = handleDone;  // BUG this below open() caused missed state==1 notifications
  re.open("POST", '/run', true);
  re.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  re.send(data);
}

// Send a remark string to /run, doing nothing with response
function sendRemark(remark){
  var data = 'id='   + encodeURIComponent(document.codeform.id.value) + '&' +
         'cuname=' + encodeURIComponent(document.codeform.cuname.value) + '&' +
         'code=&' + // empty code
         'owner=' + encodeURIComponent(document.codeform.owner.value) + '&' +
         'remark=' + encodeURIComponent(remark);
  if (document.codeform.date) {
    data +=  '&date=' + encodeURIComponent(document.codeform.date.value);
  }
  if (document.codeform.adate) {
    data +=  '&adate=' + encodeURIComponent(document.codeform.adate.value);
  }
  var req = new XMLHttpRequest();  // just local, no handler
  req.open("POST", '/run', true);
  req.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  req.send(data);
}


// Install keypress handling
// Android bluetooth cannot find a combo that works
function setupKey(editor) {
  editor.textInput.getElement().addEventListener("keypress", keyFN);
}


// handle keypresses within ACE
function keyFN(e) {
  if (e.ctrlKey && e.key=='Enter') {    
    sendCodeAce(); // focusEdit();
    //alert(e.key +' ' +e.char);
    e.preventDefault();
  }
}




/*
  editor.commands.addCommand({
    name: 'run',
    bindKey: {win: "ctrl-shift-q", mac: "ctrl-shift-q"},
    exec: function(editor) {
				        alert("yay");
				    }
	});
*/


// edit mode, send the edit data
function sendCode2(verb) {
  // loop to encode whole form, leaves extra & at end on purpose
  var data = '';
  for (var i=0; i<document.editform.elements.length; i++) {
    if (!document.editform.elements[i].disabled) {
      data += document.editform.elements[i].name + '=' +
        encodeURIComponent(document.editform.elements[i].value) + '&';
    }
  }
  
  data += 'verb=' + encodeURIComponent(verb);

  if(re.readyState > 0 && re.readyState < 4) {  // do nothing if already in flight
    return true;
  }
  re.open("POST", '/author', true);
  re.onreadystatechange = handleDone;  
  re.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  //req.overrideMimeType("text/html");
  re.send(data);
}



// get back answer
function handleDone(){
  if(re.readyState == 1) {  // starting
    document.getElementById('results').innerHTML = 'Running code...';
  }
  else if(re.readyState == 4) {  // done
    var text = re.responseText;
    document.getElementById('results').innerHTML = text;
    // move edit focus for line:xx in error message
    var line = text.indexOf('line:');
    if (line != -1) {
      var num = parseInt(text.substring(line + 5));
      if (!isNaN(num)) {
        var Range = ace.require("ace/range").Range;
        var ln = document.ace_editor.session.getLine(num-1);
        var range = new Range(num-1, ln.length, num-1, ln.length);
        document.ace_editor.selection.setSelectionRange(range, false);
      }
    }
  }
}


// Put cursor blinking at the end of the last code line
function startCursor(editor, lang) {
  var i = editor.session.getLength() - 1;
  var line;
  while (i >= 1) {
    line = editor.session.getLine(i);
    if (lang == "java" && line.indexOf('}') != -1) { i--; break; }
    if (lang == "python" && line.length >= 2) { break; }
    i--;
  }

  if (i > 0) {
    // set cursor at end of line i
    var Range = ace.require("ace/range").Range;
    line = editor.session.getLine(i);
    var range = new Range(i, line.length, i, line.length);
    editor.selection.setSelectionRange(range, false);
    editor.focus();
  }
}


// Blink the edit area cursor, handy after a button grabs focus
function focusEdit() {
  if (document.ace_editor) {
    document.ace_editor.focus();
  }
}


// GET requests to jam into divs
function getFrag(destid, path) {
  var oReq = new XMLHttpRequest();
  oReq.onload = function(e) {
    document.getElementById(destid).innerHTML=oReq.responseText;
  };
  oReq.open("GET", path);
  oReq.send();
}


// Modernized post to /runx
function sendModern(form, verb, donefn){
  var data = '';
  // hack: picks up ace_editor first as code=
  data = 'code=' + encodeURIComponent(document.ace_editor.getValue()) + '&' ;
  // loop to encode whole form, leaves extra & at end on purpose
  for (var i=0; i<form.elements.length; i++) {
    if (form.elements[i].name) {  // TODO I think ACE makes an empty elt in there
      data += form.elements[i].name + '=' +
        encodeURIComponent(form.elements[i].value) + '&';
    }
  }
  data += 'verb=' + encodeURIComponent(verb);
  //debugger;

  var oReq = new XMLHttpRequest();
  
  oReq.open('POST', '/runx');
  
  // TODO could have UI for these cases
  //oReq.addEventListener("error", transferFailed);
  //oReq.addEventListener("abort", transferCanceled);
  oReq.onload = donefn;  // this.responseText works in here
  
  oReq.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  oReq.send(data);
}


// set ace size, 80 100 120 -- as percentage
// 100 is a marker to get the default
function setFontSize(percent) {
  //alert(document.ace_editor.getFontSize());
  if (percent == 100) {
    document.ace_editor.setFontSize('small');  // this should be in sync with the .css
  }
  else {
    document.ace_editor.setFontSize(percent + '%');
  }
  document.ace_font = percent;  // note user setting for post time
}

// font size control listener
function fontChange(event) {
  setFontSize(event.target.value)
}
    
// experiment, setting width of ace div in pixels
function setWidth(width) {
  document.getElementById('ace_div').style.width = width + 'px';
  //cell.style.width = '200px';
}

