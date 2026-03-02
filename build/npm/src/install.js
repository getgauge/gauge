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

const getVersion = function (p) {
    return new Promise((resolve, reject) => {
        if (!fs.existsSync(p)) {
            reject("Unable to find package.json.");
        }
        fs.readFile(p, (err, data) => {
            if (err) {
                reject(err);
            }
            const pkg = JSON.parse(data);
            if (pkg.version) {
                resolve(pkg.version);
            } else {
                reject(new Error("Unable to find version in package.json."));
            }
        })
    });
};

const getBinaryUrl = function (version) {
    // Use x64 on Windows for ARM64 as we don't publish ARM64 builds yet
    const canonical_os = PLATFORM_MAPPING[process.platform];
    const canonical_arch = ARCH_MAPPING[canonical_os === 'windows' && process.arch === 'arm64' ? 'x64' : process.arch];
    return `${BASE_URL}v${version}/gauge-${version}-${canonical_os}.${canonical_arch}.zip`;
};

module.exports = {
    getVersion: getVersion,
    getBinaryUrl: getBinaryUrl
}