#!/usr/bin/env node

"use strict"

const install = require("./install"),
    path = require("path"),
    unzip = require('unzipper'),
    https = require('https'),
    packageJsonPath = path.join(__dirname, "..", "package.json"),
    binPath = "./bin";

var downloadFollowingRedirect = function(url, resolve, reject) {
    https.get(url, { headers: { 'accept-encoding': 'gzip,deflate' } }, res => {
        if (res.statusCode >= 300 && res.statusCode < 400) {
            downloadFollowingRedirect(res.headers.location, reject, resolve);
        } else if (res.statusCode > 400) {
            console.error(`Unable to download '${url}' : ${res.statusCode}-'${res.statusMessage}'`);
        } else {
            res.pipe(unzip.Extract({ path: path.normalize(binPath) })).on('error', reject).on('end', resolve);
        }
    });
};

var downloadAndExtract = function(version) {
    console.log(`Fetching download url for Gauge version ${version}`);
    let url = install.getBinaryUrl(version);
    console.log(`Downloading ${url} to ${binPath}`);
    return new Promise((resolve, reject) => {
        try {
            downloadFollowingRedirect(url, resolve, reject);
        } catch (error) {
            reject(error);
        }
    })
};

install.getVersion(packageJsonPath)
    .then((v) => downloadAndExtract(v.split('-')[0]))
    .catch((e) => console.error(e));