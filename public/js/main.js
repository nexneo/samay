angular.module('samay.filter', []).
filter('hoursmins', function() {
	return function(input){
		var hh = Math.floor(input.asHours()),
			mm = "0" + input.minutes();
		
		return hh+":"+mm.substr(-2);
	};
}).
filter('prettydate', function() {
	return function(input){
		return moment.unix(input).format("MMM Do, hA");
	};
}).
filter('isactive', function() {
	return function(input, compareTo){
		if(input == compareTo){
			return "active"
		}
		return ""
	};
});


angular.module('samay', ['samay.filter']).
config(function($routeProvider){
	$routeProvider.
		when("/projects", {controller:ProjectsCtrl, templateUrl:"partials/projects.html"}).
		when("/projects/:projectSha", {controller:ProjectCtrl, templateUrl:"partials/project.html"}).
		when("/projects/:projectSha/entries/:entryId", {controller:EntryCtrl, templateUrl:"partials/entry.html"}).
		otherwise({redirectTo: "/projects"});
}).
factory("projects", function($http, $route){
	var ret = [];
	function process (data) {
		var projects = [];
		for (var i = data.length - 1; i >= 0; i--) {
			project = data[i]['project'];
			project['entries'] = DecoEntries(data[i]['entries']);
			projects[i] = project;
		}
		return projects;
	}

	function DecoEntries(entries) {
		for (var i = entries.length - 1; i >= 0; i--) {
			entries[i]['type'] = 2;
			entries[i]['hours'] = moment.duration(entries[i].duration/1000000);
		}
		return entries;
	}

	$http.get("/app.json").success(function(data){
		angular.copy(process(data), ret);
		$route.reload();
	});

	return ret;
});

function EntryCtrl($scope, $routeParams, projects) {
	$scope.title = "Entry";
	$scope.projectSha = $routeParams.projectSha;
	$scope.entryId = $routeParams.entryId;
	$scope.categories = [
		{code:1, name:"Fun"},
		{code:2, name:"Work"},
		{code:0, name:"Chore"}
	];

	angular.forEach(projects, function(p){
		if (p.sha === $scope.projectSha){
			$scope.project = p;
			angular.forEach(p.entries, function(e){
				if(e.id === $scope.entryId){
					$scope.entry = e;
				}
			})
		}
	});
}

function ProjectCtrl($scope, $routeParams, projects){
	$scope.title = "Projects";
	$scope.projects = projects;
	$scope.categories = [
		{code:1, name:"Fun"},
		{code:2, name:"Work"},
		{code:0, name:"Chore"}
	];

	$scope.projectSha = $routeParams.projectSha;

	angular.forEach(projects, function(p){
		if (p.sha === $scope.projectSha){
			$scope.activeProject = p;
		}
	});

	$scope.changeType = function(entry, c){
		entry.type = c.code;
	}

	$scope.totalHours = function (){
		if($scope.activeProject === undefined){
			return moment.duration(0);
		}

		var ret = 0;
		angular.forEach($scope.activeProject.entries, function(e){
			ret += e.duration;
		});
		return moment.duration(ret/1000000);
	};

	$scope.meter = function(type) {
		if($scope.activeProject === undefined){
			return 0;
		}

		var entries = $scope.activeProject.entries;
		var sum = 0;
		angular.forEach(entries, function(e){
			if(e.type===type){sum = sum + 1;}
		});
		return Math.floor((sum/entries.length)*100);
		
	};
}

function ProjectsCtrl($scope, projects){
	$scope.title = "Projects";
	$scope.projects = projects;
}
