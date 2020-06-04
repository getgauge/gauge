#!/usr/bin/env node

"use strict"

const install = require("./install"),
    path = require("path"),
    unzip = require('unzipper'),
    request = require('superagent'),
    packageJsonPath = path.join(__dirname, "..", "package.json"),
    binPath = "./bin";


var downloadAndExtract = function(version) {
    console.log(`Fetching download url for Gauge version ${version}`);
    let url = install.getBinaryUrl(version);
    console.log(`Downloading ${url} to ${binPath}`);
    return new Promise((resolve, reject) => {
        try {
            request.get(url).pipe(unzip.Extract({ path: path.normalize(binPath) }));
            resolve();
        } catch (error) {
            reject(error);
        }
    })
}

install.getVersion(packageJsonPath)
.then((v) => downloadAndExtract(v.split('-')[0]))
.catch((e) => console.error(e));