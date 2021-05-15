#!/usr/bin/env node

"use strict"

const fs = require('fs');

const BASE_URL="https://github.com/getgauge/gauge/releases/download/",
    ARCH_MAPPING = {
        "ia32": "x86",
        "x64": "x86_64",
        "arm64": "arm64"
    },
    PLATFORM_MAPPING = {
        "darwin": "darwin",
        "linux": "linux",
        "win32": "windows"
    };
 
var getVersion = function(p) {
    return new Promise( (resolve, reject) => {
        if (!fs.existsSync(p)) {
            reject("Unable to find package.json.");
        }
        fs.readFile(p, (err, data) => {
            if(err) {
                reject(err);
            }
            resolve(JSON.parse(data).version);   
        })    
    });
}

var getBinaryUrl = function(version) {
    let os = PLATFORM_MAPPING[process.platform];
    let arch = ARCH_MAPPING[process.arch];
    return `${BASE_URL}v${version}/gauge-${version}-${os}.${arch}.zip`;
}

module.exports = {
    getVersion: getVersion,
    getBinaryUrl: getBinaryUrl
}