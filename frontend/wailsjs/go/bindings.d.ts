import * as models from './models';

export interface go {
  "main": {
    "App": {
		Close():Promise<void>
		Run(arg1:string):Promise<void>
    },
  }

}

declare global {
	interface Window {
		go: go;
	}
}
