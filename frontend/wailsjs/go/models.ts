export namespace app {
	
	export class Config {
	    anthropicKey: string;
	    openaiKey: string;
	    model: string;
	    gatewayURL: string;
	    transcribeLang: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.anthropicKey = source["anthropicKey"];
	        this.openaiKey = source["openaiKey"];
	        this.model = source["model"];
	        this.gatewayURL = source["gatewayURL"];
	        this.transcribeLang = source["transcribeLang"];
	    }
	}
	export class GatewayModel {
	    value: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new GatewayModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.label = source["label"];
	    }
	}
	export class GatewayConfig {
	    name: string;
	    url: string;
	    models: GatewayModel[];
	
	    static createFrom(source: any = {}) {
	        return new GatewayConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.models = this.convertValues(source["models"], GatewayModel);
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

