#!/usr/bin/env node

"use strict"

const install = require("./install"),
    path = require("path"),
    fs = require("fs"),
    AdmZip = require('adm-zip'),
    https = require('https'),
    packageJsonPath = path.join(__dirname, "..", "package.json"),
    binPath = "./bin";

const extractZipArchive = function (buffer) {
    return new Promise(function (resolve, reject) {
        try {
            const zip = new AdmZip(buffer);
            zip.extractAllTo(path.normalize(binPath), true, true);
            resolve();
        } catch (err) {
            reject(new Error(`Failed to extract archive from buffer: ${err.message}`));
        }
    })
};

const downloadFollowingRedirect = function (url, resolve, reject) {
    https.get(url, {headers: {'accept-encoding': 'gzip,deflate'}}, res => {
        if (res.statusCode >= 300 && res.statusCode < 400) {
            downloadFollowingRedirect(res.headers.location, resolve, reject);
            res.resume()
        } else if (res.statusCode >= 400) {
            reject(new Error(`Unable to download '${url}' : ${res.statusCode}-'${res.statusMessage}'`));
            res.resume()
        } else {
            const chunks = [];
            res
              .on('data', chunk => chunks.push(chunk))
              .on('end', () => resolve(Buffer.concat(chunks)))
              .on('error', reject);
        }
    });
};

const downloadAndExtract = function (version) {
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
      .then(extractZipArchive)
};

const removeStubBinary = () => {
    try {
        fs.unlinkSync(path.normalize(`${binPath}/gauge`));
    } catch (err) {
        if (err.code !== 'ENOENT') {
            console.error("Unable to clean up stub binary; ignoring...", err)
        }
    }
};

install.getVersion(packageJsonPath)
    .then((v) => downloadAndExtract(v.split('-')[0]))
    .catch((e) => { removeStubBinary(); throw e; });
