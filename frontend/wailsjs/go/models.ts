export namespace main {
	
	export class Instance {
	    id: string;
	    name: string;
	    type: string;
	    command: string[];
	    url: string;
	    user: string;
	    pass: string;
	    pid: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new Instance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.command = source["command"];
	        this.url = source["url"];
	        this.user = source["user"];
	        this.pass = source["pass"];
	        this.pid = source["pid"];
	        this.status = source["status"];
	    }
	}

}

