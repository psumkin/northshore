"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = require('@angular/core');
var http_1 = require('@angular/http');
require('object-assign');
var Observable_1 = require('rxjs/Observable');
require('rxjs/add/observable/interval');
require('rxjs/add/observable/throw');
require('rxjs/add/operator/catch');
require('rxjs/add/operator/map');
require('rxjs/add/operator/share');
require('rxjs/add/operator/startWith');
require('rxjs/add/operator/switchMap');
var alerts_1 = require('../alerts/alerts');
var assets_1 = require('../assets/assets');
var Blueprint = (function () {
    function Blueprint() {
    }
    return Blueprint;
}());
exports.Blueprint = Blueprint;
var APIService = (function () {
    function APIService(alertsService, assetsService, http) {
        var _this = this;
        this.alertsService = alertsService;
        this.assetsService = assetsService;
        this.http = http;
        this.blueprintsInterval = this.assetsService.asset('timers').blueprintsInterval;
        this.blueprintsUrl = this.assetsService.asset('api').blueprintsUrl;
        this.blueprints = Observable_1.Observable
            .interval(this.blueprintsInterval)
            .startWith(0)
            .switchMap(function () { return _this.http.get(_this.blueprintsUrl); })
            .map(this.extractData)
            .map(this.extendBlueprintsData)
            .share()
            .catch(function (error) { return _this.handleError(error, '#APIService.getBlueprints,#Error'); });
    }
    APIService.prototype.extendBlueprintsData = function (bps) {
        var stagesStatesBages = {};
        var filters = {
            green: ['running'],
            orange: ['new', 'created'],
            grey: ['deleted', 'paused', 'stopped'],
        };
        for (var f in filters) {
            stagesStatesBages[f] = 0;
        }
        for (var i in bps) {
            var bp = bps[i];
            bp.ui = {
                stagesStatesBages: Object.assign({}, stagesStatesBages)
            };
            for (var s in bp.stagesStates) {
                for (var f in filters) {
                    if (filters[f].indexOf(bp.stagesStates[s]) > -1) {
                        bp.ui.stagesStatesBages[f]++;
                        break;
                    }
                }
            }
        }
        return bps;
    };
    APIService.prototype.extractData = function (res) {
        var body = res.json();
        return body.data || {};
    };
    APIService.prototype.handleError = function (error, logTags) {
        console.error(logTags ? logTags : '#APIService,#Error', error);
        // handle JSONAPI Errors
        try {
            var o = error.json();
            if (o && o.errors) {
                for (var i in o.errors) {
                    this.alertsService.alertError(o.errors[i].details);
                }
            }
        }
        catch (e) {
            this.alertsService.alertError();
        }
        return Observable_1.Observable.throw(error);
    };
    /**
      @description Returns the Observable that repeats the XHR while subscribed.
     */
    APIService.prototype.getBlueprints = function () {
        return this.blueprints;
    };
    APIService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [alerts_1.AlertsService, assets_1.AssetsService, http_1.Http])
    ], APIService);
    return APIService;
}());
exports.APIService = APIService;
//# sourceMappingURL=api.js.map