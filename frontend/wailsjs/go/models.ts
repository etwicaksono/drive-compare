export namespace main {
	
	export class CompareResultItem {
	    relativePath: string;
	    name: string;
	    extension: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new CompareResultItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relativePath = source["relativePath"];
	        this.name = source["name"];
	        this.extension = source["extension"];
	        this.size = source["size"];
	    }
	}
	export class CompareResponse {
	    localOnly: CompareResultItem[];
	    driveOnly: CompareResultItem[];
	    both: CompareResultItem[];
	
	    static createFrom(source: any = {}) {
	        return new CompareResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.localOnly = this.convertValues(source["localOnly"], CompareResultItem);
	        this.driveOnly = this.convertValues(source["driveOnly"], CompareResultItem);
	        this.both = this.convertValues(source["both"], CompareResultItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
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

