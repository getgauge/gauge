package debug

var html = `
<html>

<head>
    <title>
        Gauge Debug API
    </title>
    <link href="https://getgauge.io/assets/images/favicons/favicon.ico" rel="shortcut icon" type="image/ico">
    <script lang="javascript">
        const send = () => {
            var http = new XMLHttpRequest();
            const r = document.getElementById("request").value;
            let method = "GET";
            let params = "";
            if (r == "libPathRequest") {
                params = "pluginName=" + document.getElementById("pluginName").value;
            } else if (r == "stepValueRequest") {
                params = "stepText=" + document.getElementById("stepText").value;
                params += "&hasInlineTable=" + document.getElementById("hasInlineTable").checked;
            } else if (r == "formatSpecsRequest") {
                method = "POST";
                params = "specs=" + document.getElementById("specs").value;
            } else if (r == "refactoringRequest") {
                method = "POST";
                params = "old=" + document.getElementById("old").value;
                params += "&new=" + document.getElementById("new").value;
            }
            http.open(method, "/" + r + "?" + params, true);
            http.onreadystatechange = function() {
                document.getElementById("output").innerHTML = "<pre><xmp>" + http.responseText + "</xmp></pre>";
                document.getElementById("output").hidden = false;
            }
            http.send();
        }

        const connect = () => {
            var http = new XMLHttpRequest();
            const r = document.getElementById("request").value;
            const e = document.querySelector('input[name="process"]:checked');
            if (!e) {
                alert("Please select a port");
                return;
            }
            let params = "port=" + e.value;
            http.open("POST", "/?" + params, true);
            http.onreadystatechange = function() {
                document.getElementById("connectOutput").innerHTML = "<pre>" + http.responseText + "</pre>";
                if (http.readyState == 4 && http.status == 200)
                    document.getElementById("requests").hidden = false;
            }
            http.send();
        }

        const requestChange = (v) => {
            [].slice.call(document.getElementsByClassName("options")).forEach((e) => e.hidden = true)
            const e = document.getElementById(v);
            if (e) e.hidden = false;
        }
    </script>
    <style>
        body {
            font-family: 'Open Sans', sans-serif;
            text-align: center;
            margin: 0px;
        }

        table {
            border-collapse: collapse;
            width: 100%;
        }

        th,
        td {
            border-bottom: 1px solid #ddd;
            padding: 10px;
            text-align: left;
        }

        tr:hover {
            background-color: #f5f5f5
        }

        th {
            background-color: #464545;
            color: #ffffff;
        }

        input[type=button] {
            text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.12);
            border: 1px solid rgba(0, 0, 0, 0.12);
            margin-top: 1%;
            background: #464545;
            color: #ffffff;
            font: 400 18px/1.5 "Roboto Slab", "Helvetica Neue", Helvetica, Arial, sans-serif;
            border-radius: 10px 10px 10px 10px;
            -webkit-border-radius: 10px 10px 10px 10px;
            -moz-border-radius: 10px 10px 10px 10px;
            padding-top: 0.5%;
            padding-bottom: 0.5%;
            padding-right: 1%;
            padding-left: 1%;
        }

        input[type=button]:hover {
            background: #5d5b5b;
            -moz-box-shadow: inset 0px 2px 2px 0px rgba(255, 255, 255, 0.28);
            -webkit-box-shadow: inset 0px 2px 2px 0px rgba(255, 255, 255, 0.28);
            box-shadow: inset 0px 2px 2px 0px rgba(255, 255, 255, 0.28);
            cursor: pointer;
        }

        div {
            border-radius: 20px
        }

        input[type=button]:focus {
            outline: none;
        }

        #output {
            text-align: left;
            background: #E5E9EB;
            border-radius: 20px;
            padding: 1%;
            margin-top: 1%;
            overflow: auto;
        }

        #connectOutput {
            margin-top: 0.5%;
        }

        th:first-child {
            border-top-left-radius: 20px;
        }

        th:last-child {
            border-top-right-radius: 20px;
        }

        .options {
            margin-top: 1%;
        }

        textarea {
            width: 30%;
            height: 50px;
            font-size: 14px;
            font-family: "Courier New", Courier, monospace;
        }

        .menu {
            width: 100%;
            height: 65px;
            background: #464545;
            z-index: 100;
            -webkit-touch-callout: none;
            -webkit-user-select: none;
            -moz-user-select: none;
            -ms-user-select: none;
            border-bottom: 2px solid #7b7b7b;
            margin-bottom: 1%;
            border-radius: 0px;
        }

        .back {
            position: absolute;
            width: 90px;
            height: 50px;
            top: 5px;
            left: 0px;
            color: #000000;
            line-height: 45px;
            font-size: 40px;
            padding-left: 10px;
            cursor: pointer;
            transition: .15s all;
            text-decoration: none;
        }

        .back img {
            position: absolute;
            top: 10px;
            left: 50px;
            height: 35px;
            margin-left: 15px;
        }

        .beta {
            position: relative;
            background: #f5c10e;
            color: #000;
            padding: 2px 6px;
            font-size: 12px;
            text-transform: uppercase;
            top: 6px;
            left: 177px;
            border-radius: 6px;
        }

        .back:active {
            background: rgba(0, 0, 0, 0.15);
        }

        select {
            height: 35px;
            width: 200px;
            font-size: 14px;
            margin-right: 2%;
        }

        select:focus {
            outline: none;
        }

        input[type="text"] {
            height: 30px;
            width: 300px;
            font-size: 14px;
            font-family: "Courier New", Courier, monospace;
            margin-right: 50px;
        }

        #requests {
            border: 1px solid #9a988f;
            padding-bottom: 1%;
            margin-top: 1%;
        }
    </style>
</head>

<body>
    <div class="menu">
        <a href="http://getgauge.io" target="_blank" class="back"><img src="https://getgauge.io/assets/images/Gauge_logo.svg" draggable="false" /><span class="beta">BETA</span></a>
    </div>
    <div style="margin-right: 1%; margin-left: 1%;">
        <div style="border: 1px solid #9a988f;padding-bottom: 0.5%;">
            <table>
                <tr>
                    <th></th>
                    <th>PID</th>
                    <th>Port</th>
                    <th>CWD</th>
                </tr>
                {{range .}}
                <tr>
                    <td><input type="radio" name="process" value="{{.Port}}"></td>
                    <td>{{.Pid}}</td>
                    <td>{{.Port}}</td>
                    <td>{{.Cwd}}</td>
                </tr>
                {{end}}
            </table>
            <input type="button" value="Connect" onclick="connect()" />
            <div id="connectOutput"></div>
        </div>
        <div id="requests" hidden>
            <select id="request" onchange="requestChange(this.value)">
                <option value="projectRootRequest">ProjectRootRequest</option>
                <option value="installationRootRequest">InstallationRootRequest</option>
                <option value="libPathRequest">LibPathRequest</option>
                <option value="allStepsRequest">AllStepsRequest</option>
                <option value="allConceptsRequest">AllConceptsRequest</option>
                <option value="specsRequest">SpecsRequest</option>
                <option value="stepValueRequest">StepValueRequest</option>
                <option value="formatSpecsRequest">FormatSpecsRequest</option>
                <option value="refactoringRequest">RefactoringRequest</option>
            </select>

            <input type="button" value="Send" onclick="send()" />

            <div id="stepValueRequest" class="options" hidden>
                <textarea id="stepText" placeholder="Enter step text..."></textarea><br><br>
                <label for="hasInlineTable">Has inline table</label>
                <input type="checkbox" id="hasInlineTable" />
            </div>

            <div id="libPathRequest" class="options" hidden>
                <label for="pluginName">Enter plugin name: </label>
                <input type="text" id="pluginName" />
            </div>

            <div id="formatSpecsRequest" class="options" hidden>
                <label for="specs">Enter spec file(s) path</label>
                <textarea id="specs" placeholder="specs/example.spec&#10;specs/example1.spec"></textarea>
            </div>

            <div id="refactoringRequest" class="options" hidden>
                <label for="old">Old Step: </label>
                <input type="text" id="old" />
                <label for="new">New Step: </label>
                <input type="text" id="new" />
            </div>
        </div>
        <div id="output" hidden></div>
    </div>
</body>

</html>
`
