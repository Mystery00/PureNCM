export namespace config {
	
	export class Config {
	    outputDir: string;
	    filenamePattern: string;
	    copyLrc: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outputDir = source["outputDir"];
	        this.filenamePattern = source["filenamePattern"];
	        this.copyLrc = source["copyLrc"];
	    }
	}

}

