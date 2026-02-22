export namespace main {
	
	export class Config {
	    anthropicKey: string;
	    openaiKey: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.anthropicKey = source["anthropicKey"];
	        this.openaiKey = source["openaiKey"];
	        this.model = source["model"];
	    }
	}

}

