function SamayHomeCtrl($scope, $http){
	$http.get("/app.json").success(function(data){
		var projects = [];
		for (var i = data.length - 1; i >= 0; i--) {
			project = data[i]['project'];
			project['entries'] = data[i]['entries'];
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
}
