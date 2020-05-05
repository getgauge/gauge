"use strict"

const expect = require('chai').expect,
    fs = require('fs'),
    sinon = require('sinon');

var subject = require("../src/install");

describe("getVersion", () => {
    it("should get the version in package.json", async () => {
        const dummyPath = "/foo/bar";
        sinon.stub(fs, "existsSync").returns(true);
        sinon.stub(fs, "readFile").withArgs(dummyPath).yields(undefined, '{"version": "x.y.z"}');
        
        expect(await subject.getVersion(dummyPath)).equal("x.y.z");
    });
});

describe("getBinaryUrl", () => {
    it("should return platform specific URL", async () => {
        let originalPlatform = Object.getOwnPropertyDescriptor(process, 'platform');;
        let originalArch = Object.getOwnPropertyDescriptor(process, 'arch');;
        Object.defineProperty(process, 'platform', { value: "win32" });
        Object.defineProperty(process, 'arch', { value: "ia32" });

        expect(await subject.getBinaryUrl("1.0.0")).equals("https://github.com/getgauge/gauge/releases/download/v1.0.0/gauge-1.0.0-windows.x86.zip");

        Object.defineProperty(process, 'platform', originalPlatform);
        Object.defineProperty(process, 'arch', originalArch);
    });

});


