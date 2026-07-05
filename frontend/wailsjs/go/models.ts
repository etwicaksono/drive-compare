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
	    sourceAOnly: CompareResultItem[];
	    sourceBOnly: CompareResultItem[];
	    both: CompareResultItem[];
	
	    static createFrom(source: any = {}) {
	        return new CompareResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceAOnly = this.convertValues(source["sourceAOnly"], CompareResultItem);
	        this.sourceBOnly = this.convertValues(source["sourceBOnly"], CompareResultItem);
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
	export class DuplicateGroupResponse {
	    baseName: string;
	    extension: string;
	    relativePath: string;
	    files: CompareResultItem[];
	
	    static createFrom(source: any = {}) {
	        return new DuplicateGroupResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseName = source["baseName"];
	        this.extension = source["extension"];
	        this.relativePath = source["relativePath"];
	        this.files = this.convertValues(source["files"], CompareResultItem);
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
	export class DuplicatePatternInfo {
	    id: string;
	    label: string;
	    example: string;
	
	    static createFrom(source: any = {}) {
	        return new DuplicatePatternInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.example = source["example"];
	    }
	}
	export class DuplicateScanResponse {
	    groups: DuplicateGroupResponse[];
	    totalGroups: number;
	    totalFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new DuplicateScanResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groups = this.convertValues(source["groups"], DuplicateGroupResponse);
	        this.totalGroups = source["totalGroups"];
	        this.totalFiles = source["totalFiles"];
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

}

