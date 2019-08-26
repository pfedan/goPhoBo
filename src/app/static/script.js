var modal = document.getElementById('myModal');
var modalImg = document.getElementById("img01");

var xmlhttp = new XMLHttpRequest();
var xmlhttpRefresh = new XMLHttpRequest();

var cntPhotos = 0;
var currentState = '';

var lastCountdownStart = 0;
var tCal = 0;

xmlhttp.onreadystatechange = getNewImageList;
xmlhttp.open("GET", "../images", true);
xmlhttp.send();

autoRefresh();

// Get the <span> element that closes the modal
var span = document.getElementsByClassName("close")[0];
span.onclick = function () {
    modal.style.display = "none";
}
modal.onclick = function () {
    modal.style.display = "none";
}

document.addEventListener("keypress", function onEvent(event) {
    if (event.key == "s") {
        var statusBox = document.getElementById('status')
        if (statusBox.style.display == "block") {
            statusBox.style.display = "none";
        } else {
            statusBox.style.display = "block"
        }
    } else if (event.key == "a") {
        var stateBox = document.getElementById('currentState')
        if (stateBox.style.display == "block") {
            stateBox.style.display = "none";
        } else {
            stateBox.style.display = "block"
        }
    } else if (event.key == "p") {
        showCountDown()
    } else if (event.key == "c") {
        var d = new Date();
        tCal = d.getTime() - lastCountdownStart - 4000;
        console.log("tCal: " + tCal);
    } else if (event.key == "!") {
        if (confirm("This will delete ALL photos. Please confirm to proceed.")) {
            loadXMLDoc("../deleteAll", function () { });
            location.reload();
        }
    }
});

function showCountDown() {
    var d = new Date();
    lastCountdownStart = d.getTime();
    var divCD = document.getElementById("countdown");
    var divCDNum = document.getElementById("countdown_content");
    divCDNum.innerHTML = "4";
    divCD.style.display = "block";
    setTimeout(function () { showCountdownNumber(3) }, 1000);
    setTimeout(function () { showCountdownNumber(2) }, 2000);
    setTimeout(function () { showCountdownNumber(1) }, 3000);
    setTimeout(function () { showCountdownNumber("Smile!") }, 4000);
    setTimeout(function () { loadXMLDoc("../doPhoto", function () { }) }, 4000 - tCal);
    setTimeout(function () { divCD.style.display = "none" }, 5000);
}
function showCountdownNumber(num) {
    var divCDNum = document.getElementById("countdown_content");
    divCDNum.innerHTML = num.toString();
}

function loadXMLDoc(url, cfunc) {
    //Code to catch modern browsers
    if (window.XMLHttpRequest) {
        xmlhttpRefresh = new XMLHttpRequest();
    }

    //Code to catch crap browsers
    else {
        xmlhttpRefresh = new ActiveXObject("Microsoft.XMLHTTP");
    }

    //Set up
    xmlhttpRefresh.onreadystatechange = cfunc;
    xmlhttpRefresh.open("GET", url, true);
    xmlhttpRefresh.send();
}

function getNewImageList() {
    if (this.readyState == 4 && this.status == 200) {
        var res = JSON.parse(this.responseText);
        if ("imageFiles" in res && res.imageFiles.length != cntPhotos && currentState == 'home') {
            console.log(res.imageFiles[0]);
            makeImageView(res.imageFiles);
        }
        //document.getElementById('json').innerHTML = JSON.stringify(res, undefined, 2);
    }
}

function makeImageView(list) {
    //document.getElementById("gallery").innerHTML = '';
    var i;
    for (i = cntPhotos; i < list.length; i++) {
        var node = document.createElement("DIV");
        node.innerHTML = '<div class="polaroid">' + //'<a href="../img/' + list[i] + '">' +
            '<img class="myImg" src="../img/small/' + list[i] + '" alt="' + list[i] + '">' + //'</a>' +
            '<div class="container">' +
            //'<p>' + list[i] + '</p>' +
            '</div>' +
            '</div>';

        document.getElementById("gallery").appendChild(node);
        var polaroids = document.getElementsByClassName("polaroid");
        polaroids[polaroids.length - 1].style.transform = "rotate(" + (Math.floor(Math.random() * 20) - 10).toString() + "deg)";
    }

    var imgList = document.getElementsByClassName('myImg');
    for (i = cntPhotos; i < imgList.length; i++) {
        imgList[i].onclick = function () {
            modal.style.display = "flex";
            modalImg.src = this.src.replace('small/', '');
        }
    }
    cntPhotos = list.length;
}

function autoRefresh() {
    console.log("called autoRefresh")
    var target = document.getElementById('json');
    var url = '../status';

    var doRefresh = function () {
        xmlhttp.open("GET", "../images", true);
        xmlhttp.send();
        loadXMLDoc(url, function () {
            if (xmlhttpRefresh.readyState == 4 && xmlhttpRefresh.status == 200) {
                var res = JSON.parse(xmlhttpRefresh.responseText);
                target.innerHTML = JSON.stringify(res, undefined, 2);
                currentState = res.currentState;
                document.getElementById('currentState').innerHTML = currentState;
            };
        });
    }
    setInterval(doRefresh, 350);
}