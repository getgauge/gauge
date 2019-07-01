#!/usr/bin/env node

"use strict"

const fs = require('fs'),
    request = require('request');

const BASE_URL="https://api.github.com/repos/getgauge/gauge/releases",
    ARCH_MAPPING = {
        "ia32": "x86",
        "x64": "x86_64"
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

var getReleaseURL = function(version) {
    return `${BASE_URL}/tags/v${version}`;
}

var getBinaryUrl = function(version) {
    return new Promise((resolve, reject) => {
        let url = getReleaseURL(version);
        
        let os = PLATFORM_MAPPING[process.platform];
        let arch = ARCH_MAPPING[process.arch];

        request.get(url, { headers: {'user-agent': 'node.js'}, json: true}, (err, res, data) => {
            try {
                if( err ) reject(new Error(err));
                if ( res && res.statusCode >= 400 ) reject(new Error (res.body.message));
                if (!data.assets) reject(new Error('Please check your internet connection. Also ensure that you are not behind any firewall.'))
                for (const key in data.assets) {
                    if (data.assets.hasOwnProperty(key)) {
                        const a = data.assets[key];
                        if(a.browser_download_url.indexOf(`${os}.${arch}.zip`) >= 0) {
                            resolve(a.browser_download_url);
                        }
                    }
                }
                reject(new Error(`No download link found for version ${version}, OS: ${os}, Arch: ${arch}`));
            } catch (error) {
                reject(error);
            }
        });
    });
}

module.exports = {
    getVersion: getVersion,
    getReleaseURL: getReleaseURL,
    getBinaryUrl: getBinaryUrl
}