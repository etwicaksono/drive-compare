export namespace main {
	
	export class DriveFolder {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new DriveFolder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

