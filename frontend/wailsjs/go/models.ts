export namespace main {
	
	export class UploadRecord {
	    fileName: string;
	    size: number;
	    timestamp: string;
	    savePath: string;
	
	    static createFrom(source: any = {}) {
	        return new UploadRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.size = source["size"];
	        this.timestamp = source["timestamp"];
	        this.savePath = source["savePath"];
	    }
	}

}

