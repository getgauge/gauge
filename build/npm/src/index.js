#!/usr/bin/env node

"use strict"

const install = require("./install"),
    path = require("path"),
    request = require('request'),
    unzip = require('unzipper'),
    fs = require('fs'),
    packageJsonPath = path.join(__dirname, "..", "package.json"),
    binPath = "./bin";


var downloadAndExtract = async function(version) {
    console.log(`Fetching download url for Gauge version ${version}`);
    let url = await install.getBinaryUrl(version);
    let gaugeExecutable = process.platform === "win32" ? "gauge.exe" : "gauge"
    console.log(`Downloading ${url} to ${binPath}`);
    unzip.Open.url(request, url).then((d) => {
        return new Promise((resolve, reject) => {
            d.files[0].stream()
            .pipe(fs.createWriteStream(path.join(binPath, gaugeExecutable)))
            .on('error',reject)
            .on('finish',resolve)
        });
    });
}

install.getVersion(packageJsonPath)
.then((v) => downloadAndExtract(v.split('-')[0]))
.catch((e) => console.error(e));