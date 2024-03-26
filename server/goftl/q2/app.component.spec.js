System.register(['./app.component', 'angular2/testing', 'angular2/platform/browser'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var app_component_1, testing_1, browser_1;
    return {
        setters:[
            function (app_component_1_1) {
                app_component_1 = app_component_1_1;
            },
            function (testing_1_1) {
                testing_1 = testing_1_1;
            },
            function (browser_1_1) {
                browser_1 = browser_1_1;
            }],
        execute: function() {
            ////////  SPECS  /////////////
            /// Delete this
            testing_1.describe('Smoke test', function () {
                testing_1.it('should run a passing test', function () {
                    testing_1.expect(true).toEqual(true, 'should pass');
                });
            });
            testing_1.describe('AppComponent with new', function () {
                testing_1.it('should instantiate component', function () {
                    testing_1.expect(new app_component_1.AppComponent()).toBeDefined('Whoopie!');
                });
            });
            testing_1.describe('AppComponent with TCB', function () {
                testing_1.it('should instantiate component', testing_1.async(testing_1.inject([testing_1.TestComponentBuilder], function (tcb) {
                    tcb.createAsync(app_component_1.AppComponent).then(function (fixture) {
                        testing_1.expect(fixture.componentInstance instanceof app_component_1.AppComponent).toBe(true, 'should create AppComponent');
                    });
                })));
                testing_1.it('should have expected <h1> text', testing_1.async(testing_1.inject([testing_1.TestComponentBuilder], function (tcb) {
                    tcb.createAsync(app_component_1.AppComponent).then(function (fixture) {
                        // fixture.detectChanges();  // would need to resolve a binding but we don't have a binding
                        var h1 = fixture.debugElement.query(function (el) { return el.name === 'h1'; }).nativeElement; // it works
                        h1 = fixture.debugElement.query(browser_1.By.css('h1')).nativeElement; // preferred
                        testing_1.expect(h1.innerText).toMatch(/angular 2 app/i, '<h1> should say something about "Angular 2 App"');
                    });
                })));
            });
        }
    }
});
//# sourceMappingURL=app.component.spec.js.map