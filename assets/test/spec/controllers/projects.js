'use strict';

describe('Controller: ProjectsController', function () {

  // Angular is refusing to recognize the HawtioNav stuff
  // when testing even though its being loaded
   beforeEach(module(function ($provide) {
    $provide.provider("HawtioNavBuilder", function() {
      function Mocked() {}
      this.create = function() {return this;};
      this.id = function() {return this;};
      this.title = function() {return this;};
      this.template = function() {return this;};
      this.isSelected = function() {return this;};
      this.href = function() {return this;};
      this.page = function() {return this;};
      this.subPath = function() {return this;};
      this.build = function() {return this;};
      this.join = function() {return "";};
      this.$get = function() {return new Mocked();};
    });

    $provide.factory("HawtioNav", function(){
      return {add: function() {}};
    });
  }));

  angular.module('openshiftConsole').config(function(AuthServiceProvider) {
    AuthServiceProvider.LoginService('RedirectLoginService');
    AuthServiceProvider.UserStore('MemoryUserStore');
  });

  // load the controller's module
  beforeEach(module('openshiftConsole'));

  var ProjectsController,
    scope,
    timeout;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $timeout, $rootScope, MemoryUserStore) {
    // Set up a stub user
    MemoryUserStore.setToken("myToken");
    MemoryUserStore.setUser({metadata: {name: "My User"}});

    scope = $rootScope.$new();
    timeout = $timeout;

    ProjectsController = $controller('ProjectsController', {
      $scope: scope,
      DataService: {
        list: function(type, context, callback, opts) {
          // TODO return mocked project data
          callback({by: function(){return {}}});
        }
      }
    });
  }));

  it('should set the user', function () {
    // Flush async withUser and DataService calls
    timeout.flush();
    expect(scope.user).toBeDefined();
    expect(scope.user).not.toBe(null);
  });

  it('should create the empty project list', function () {
    expect(scope.projects).toBeDefined();
    expect(scope.projects).not.toBe(null);
  });
});
