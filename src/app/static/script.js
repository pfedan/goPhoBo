var modal = document.getElementById('myModal');
var modalImg = document.getElementById("img01");

var xmlhttp = new XMLHttpRequest();
var xmlhttpRefresh = new XMLHttpRequest();

var cntPhotos = 0;
var latestImageName = '';
var currentState = '';
var newState = '';

var lastCountdownStart = 0;
var tCal = 0;

xmlhttp.onreadystatechange = getNewImageList;
xmlhttp.open("GET", "../images", true);
xmlhttp.send();

autoRefresh();

var closeModal = document.getElementById("closeModal");
closeModal.onclick = function () {
    if (currentState == "home") {
        hideElement("myModal")
    }
}
modal.onclick = function () {
    if (currentState == "home") {
        hideElement("myModal")
    }
}

document.getElementById("btnAccept").onclick = function () {
    acceptPhoto();
}

document.getElementById("btnDismiss").onclick = function () {
    deletePhoto();
}

function acceptPhoto() {
    hideElement("myModal");
    loadXMLDoc("../acceptPhoto", function () { });
}

function deletePhoto() {
    hideElement("myModal");
    loadXMLDoc("../deletePhoto", function () { });
}

function beginSmile() {
    loadXMLDoc("../beginSmile", function () { });
}

function toggleElementVisibility(element) {
    var mNode = document.getElementById(element)
    if (mNode.style.display == "block") {
        mNode.style.display = "none";
    } else {
        mNode.style.display = "block"
    }
}

function showElement(element) {
    var mNode = document.getElementById(element)
    mNode.style.display = "block"
}

function hideElement(element) {
    var mNode = document.getElementById(element)
    mNode.style.display = "none"
}

document.addEventListener("keypress", function onEvent(event) {
    switch (event.key) {
        case "1":
            if (currentState == "home") {
                showCountDownAndMakePhoto()
            }
            else if (currentState == "decide") {
                acceptPhoto();
            }
            break;
        case "2":
            if (currentState == "home") {
                beginSmile()
            }
            else if (currentState == "decide") {
                deletePhoto();
            }
            break;
        case "s":
            toggleElementVisibility('status');
            break;
        case "a":
            toggleElementVisibility('currentState')
            break;
        case "y":
            acceptPhoto();
            break;
        case "n":
            deletePhoto();
            break;
        case "p":
            showCountDownAndMakePhoto();
            break;
        case "c":
            var d = new Date();
            tCal = d.getTime() - lastCountdownStart - 4000;
            if (Math.abs(tCal) > 2000) {
                tCal = 0;
            }
            console.log("tCal: " + tCal);
            break;
        case "!":
            if (confirm("This will delete ALL photos. Please confirm to proceed.")) {
                loadXMLDoc("../deleteAll", function () { });
                location.reload();
            }
            break;
        default:
            console.log("Unregistered key-event:\'" + event.key + "\'.")
    }
});

function handleRemoteCommand(cmd) {
    loadXMLDoc("../status/remoteCommand/nothing", function () { })
    switch (cmd) {
        case "nothing":
            break;
        case "doPhoto":
            showCountDownAndMakePhoto();
            break;
        case "beginSmile":
            beginSmile();
            break;
        case "acceptPhoto":
            acceptPhoto();
            break;
        case "deletePhoto":
            deletePhoto();
            break;
    }
}

function handleNewState() {
    if (newState == currentState) { return; }

    switch (newState) {
        case "home":
            hideElement("decisionInstruction");
            break;
        case "decide":
            loadXMLDoc("../images",
                function () {
                    getNewImageList();
                    showElement("decisionInstruction");
                    showCountdownNumber("");
                    hideElement("countdown");
                    modal.style.display = "flex";
                    modalImg.src = "../img/" + latestImageName;
                }
            )
            break;
        default:

            break;
    }
    currentState = newState;
}

function showCountDownAndMakePhoto() {
    var divCD = document.getElementById("countdown");
    if (divCD.style.display == "block") { return; }

    var d = new Date();
    lastCountdownStart = d.getTime();
    divCD.style.display = "block";
    setTimeout(function () { showCountdownNumber(4) }, 0);
    setTimeout(function () { showCountdownNumber(3) }, 1000);
    setTimeout(function () { showCountdownNumber(2) }, 2000);
    setTimeout(function () { showCountdownNumber(1) }, 3000);
    setTimeout(function () { showCountdownNumber("Smile") }, 4000);
    setTimeout(function () { loadXMLDoc("../doPhoto", function () { }) }, 4000 - tCal);
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

        if (Array.isArray(res.imageFiles)) {
            latestImageName = res.imageFiles.slice(-1)[0];
        }

        if (Array.isArray(res.imageFiles) && currentState == 'home') {
            makeImageView(res.imageFiles);
        }
    }
}

function makeImageView(list) {
    if (list.length == cntPhotos) { return; }

    var i;
    for (i = cntPhotos; i < list.length; i++) {
        var node = document.createElement("DIV");
        node.innerHTML = '<div class="polaroid">' +
            '<img class="myImg" src="../img/small/' + list[i] + '" alt="' + list[i] + '">' +
            '<div class="container">' +
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
                newState = res.currentState;
                document.getElementById('currentState').innerHTML = newState;

                handleRemoteCommand(res.remoteCommand);
                handleNewState();
            };
        });
    }
    setInterval(doRefresh, 350);
}