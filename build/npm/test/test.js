"use strict"

const expect  = require('chai').expect,
    fs = require('fs'),
    sinon = require('sinon'),
    request = require('request');

var subject = require("../src/install");

describe("getVersion", () => {
    it("should get the version in package.json", async () => {
        const dummyPath = "/foo/bar";
        sinon.stub(fs, "existsSync").returns(true);
        sinon.stub(fs, "readFile").withArgs(dummyPath).yields(undefined, '{"version": "x.y.z"}');
        
        expect(await subject.getVersion(dummyPath)).equal("x.y.z");
    });
});

describe("getReleaseURl", () => {
    it("should get url for version", () => {
        let version = "1.0.0";
        let url = subject.getReleaseURL(version);

        expect(url).equals("https://api.github.com/repos/getgauge/gauge/releases/tags/v" + version);
    })
});

describe("getBinaryMeta", () => {
    it("should fetch platform specific metadata", async () => {
        let url = subject.getReleaseURL();
        let response = {
            assets: {
                0 : {
                    browser_download_url: "https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-darwin.x86.zip"
                },
                1 : {
                    browser_download_url: "https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-darwin.x86_64.zip"
                },
                2 : {
                    browser_download_url: "https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-linux.x86.zip"
                },
                3 : {
                    browser_download_url: "https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-linux.x86_64.zip"
                },
                4 : {
                    browser_download_url: "https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-windows.x86.zip"
                },
                5 : {
                    browser_download_url: "https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-windows.x86_64.zip"
                },

            }
        }
        sinon.stub(request, 'get').yields(undefined, undefined, response);

        let originalPlatform = Object.getOwnPropertyDescriptor(process, 'platform');;
        let originalArch = Object.getOwnPropertyDescriptor(process, 'arch');;
        Object.defineProperty(process, 'platform', { value: "win32" });
        Object.defineProperty(process, 'arch', { value: "ia32" });
        
        expect(await subject.getBinaryUrl()).equals("https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-windows.x86.zip");

        Object.defineProperty(process, 'platform', originalPlatform);
        Object.defineProperty(process, 'arch', originalArch);
    });
});


