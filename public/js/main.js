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
});



angular.module('samay', ['samay.filter']);


function SamayHomeCtrl($scope, $http){
	function DecoEntries(entries) {
		for (var i = entries.length - 1; i >= 0; i--) {
			entries[i]['hours'] = moment.duration(entries[i].duration/1000000);
		}
		return entries;
	}

	$scope.notEmpty = function (str) {
		return str != "";
	}

	$http.get("/app.json").success(function(data){
		var projects = [];
		for (var i = data.length - 1; i >= 0; i--) {
			project = data[i]['project'];
			project['entries'] = DecoEntries(data[i]['entries']);
			projects[i] = project;
		}
		$scope.projects = projects;
		$scope.activate($scope.projects[0])
	});

	$scope.activate = function(project) {
		angular.forEach($scope.projects, function(p){
			p.activecls = "";
		})
		project.activecls = "active";
		$scope.activeProject = project;
	};

	$scope.totalHours = function (){
		if($scope.activeProject === undefined){
			return moment.duration(0);
		}
		
		var ret = 0;
		angular.forEach($scope.activeProject.entries, function(e){
			ret += e.duration;
		});
		return moment.duration(ret/1000000)
	}
}
