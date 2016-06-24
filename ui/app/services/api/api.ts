import { Injectable } from '@angular/core';
import { Http, Response } from '@angular/http';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/observable/interval';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/share';
import 'rxjs/add/operator/startWith';
import 'rxjs/add/operator/switchMap';

import { AlertsService } from '../alerts/alerts';
import { AssetsService } from '../assets/assets';

export class Blueprint {
  name: string;
  provisioner: string; //Provisioner type (docker/...)
  stages: Object[];
  state: string;
  stateStages: Object[];
  type: string; //Type of blueprint (pipeline/application)
  version: string;
}

@Injectable()
export class APIService {

  private blueprints: Observable<Blueprint[]>;
  private blueprintsInterval = this.assetsService.asset('timers').blueprintsInterval;
  private blueprintsUrl = this.assetsService.asset('api').blueprintsUrl;

  constructor(
    private alertsService: AlertsService,
    private assetsService: AssetsService,
    private http: Http
  ) {

    this.blueprints = Observable
      .interval(this.blueprintsInterval)
      .startWith(0)
      .switchMap(() => this.http.get(this.blueprintsUrl))
      .map(this.extractData)
      .share()
      .catch(error => this.handleError(error, '#APIService.getBlueprints,#Error'));

  }

  private extractData(res: Response) {
    let body = res.json();
    return body.data || {};
  }

  private handleError(error: any, logTags?: string) {
    console.error(logTags ? logTags : '#APIService,#Error', error);
    // handle JSONAPI Errors
    try {
      let o = error.json()
      if (o && o.errors) {
        for (let i in o.errors) {
          this.alertsService.alertError(o.errors[i].details);
        }
      }
    } catch (e) {
      this.alertsService.alertError();
    }

    return Observable.throw(error);
  }

  /**
    @description Returns the Observable that repeats the XHR while subscribed.
   */
  getBlueprints(): Observable<Blueprint[]> {
    return this.blueprints;
  }

}
