export namespace chat {
	
	export class ChatMessage {
	    id: string;
	    sessionId: string;
	    role: string;
	    content: string;
	    // Go type: time
	    timestamp: any;
	    generatedTasks?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.generatedTasks = source["generatedTasks"];
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
	export class ChatSession {
	    id: string;
	    workspaceId: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

export namespace ide {
	
	export class Workspace {
	    version: string;
	    projectRoot: string;
	    displayName: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    lastOpenedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.projectRoot = source["projectRoot"];
	        this.displayName = source["displayName"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.lastOpenedAt = this.convertValues(source["lastOpenedAt"], null);
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
	export class WorkspaceSummary {
	    id: string;
	    displayName: string;
	    projectRoot: string;
	    // Go type: time
	    lastOpenedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.projectRoot = source["projectRoot"];
	        this.lastOpenedAt = this.convertValues(source["lastOpenedAt"], null);
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

export namespace main {
	
	export class ChatResponseDTO {
	    message: chat.ChatMessage;
	    generatedTasks: orchestrator.Task[];
	    understanding: string;
	    conflicts?: meta.PotentialConflict[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatResponseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = this.convertValues(source["message"], chat.ChatMessage);
	        this.generatedTasks = this.convertValues(source["generatedTasks"], orchestrator.Task);
	        this.understanding = source["understanding"];
	        this.conflicts = this.convertValues(source["conflicts"], meta.PotentialConflict);
	        this.error = source["error"];
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
	export class LLMConfigDTO {
	    kind: string;
	    model: string;
	    baseUrl: string;
	    systemPrompt: string;
	    hasApiKey: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LLMConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.model = source["model"];
	        this.baseUrl = source["baseUrl"];
	        this.systemPrompt = source["systemPrompt"];
	        this.hasApiKey = source["hasApiKey"];
	    }
	}
	export class ModelOptionDTO {
	    id: string;
	    name: string;
	    group: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelOptionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.group = source["group"];
	    }
	}
	export class ToolOptionDTO {
	    id: string;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolOptionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}

}

export namespace meta {
	
	export class PotentialConflict {
	    file: string;
	    tasks: string[];
	    warning: string;
	
	    static createFrom(source: any = {}) {
	        return new PotentialConflict(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.tasks = source["tasks"];
	        this.warning = source["warning"];
	    }
	}

}

export namespace orchestrator {
	
	export class Artifacts {
	    files?: string[];
	    logs?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Artifacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = source["files"];
	        this.logs = source["logs"];
	    }
	}
	export class Attempt {
	    id: string;
	    taskId: string;
	    status: string;
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    finishedAt?: any;
	    errorSummary?: string;
	
	    static createFrom(source: any = {}) {
	        return new Attempt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.taskId = source["taskId"];
	        this.status = source["status"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.finishedAt = this.convertValues(source["finishedAt"], null);
	        this.errorSummary = source["errorSummary"];
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
	export class BacklogItem {
	    id: string;
	    taskId: string;
	    type: string;
	    title: string;
	    description: string;
	    priority: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    resolvedAt?: any;
	    resolution?: string;
	    metadata?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new BacklogItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.taskId = source["taskId"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.priority = source["priority"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.resolvedAt = this.convertValues(source["resolvedAt"], null);
	        this.resolution = source["resolution"];
	        this.metadata = source["metadata"];
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
	export class Pool {
	    id: string;
	    name: string;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new Pool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class PoolSummary {
	    poolId: string;
	    running: number;
	    queued: number;
	    failed: number;
	    total: number;
	    counts: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new PoolSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.poolId = source["poolId"];
	        this.running = source["running"];
	        this.queued = source["queued"];
	        this.failed = source["failed"];
	        this.total = source["total"];
	        this.counts = source["counts"];
	    }
	}
	export class RunnerSpec {
	    maxLoops?: number;
	    workerKind?: string;
	
	    static createFrom(source: any = {}) {
	        return new RunnerSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxLoops = source["maxLoops"];
	        this.workerKind = source["workerKind"];
	    }
	}
	export class SuggestedImpl {
	    language?: string;
	    filePaths?: string[];
	    constraints?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SuggestedImpl(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.filePaths = source["filePaths"];
	        this.constraints = source["constraints"];
	    }
	}
	export class Task {
	    id: string;
	    title: string;
	    status: string;
	    poolId: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    // Go type: time
	    startedAt?: any;
	    // Go type: time
	    doneAt?: any;
	    description?: string;
	    dependencies?: string[];
	    parentId?: string;
	    wbsLevel?: number;
	    phaseName?: string;
	    milestone?: string;
	    sourceChatId?: string;
	    acceptanceCriteria?: string[];
	    attemptCount?: number;
	    // Go type: time
	    nextRetryAt?: any;
	    suggestedImpl?: SuggestedImpl;
	    artifacts?: Artifacts;
	    runner?: RunnerSpec;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.status = source["status"];
	        this.poolId = source["poolId"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.doneAt = this.convertValues(source["doneAt"], null);
	        this.description = source["description"];
	        this.dependencies = source["dependencies"];
	        this.parentId = source["parentId"];
	        this.wbsLevel = source["wbsLevel"];
	        this.phaseName = source["phaseName"];
	        this.milestone = source["milestone"];
	        this.sourceChatId = source["sourceChatId"];
	        this.acceptanceCriteria = source["acceptanceCriteria"];
	        this.attemptCount = source["attemptCount"];
	        this.nextRetryAt = this.convertValues(source["nextRetryAt"], null);
	        this.suggestedImpl = this.convertValues(source["suggestedImpl"], SuggestedImpl);
	        this.artifacts = this.convertValues(source["artifacts"], Artifacts);
	        this.runner = this.convertValues(source["runner"], RunnerSpec);
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

