package routes

import (
	routeapi "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRoute returns a route object,
// It takes the service to route as a parameter
func NewRoute(routeName string, serviceName string, clusterName string, routeType string) *routeapi.Route {
	r := routeapi.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: routeName,
			Labels: map[string]string{
				"oshinko-cluster": clusterName,
				"oshinko-type": routeType,
			},
		},
		Spec: routeapi.RouteSpec{
			To: routeapi.RouteTargetReference{
				Name: serviceName,
			},
		},
	}

	//r.Kind = "Route"
	//r.APIVersion = "v1"
	//r.SetName(name)
	//r.SetNamespace(namespace)
	//rSpec := api.RouteSpec{}
	//api.Route{
	//		ObjectMeta: kapi.ObjectMeta{
	//			Name: routeName,
	//		},
	//		Spec: api.RouteSpec{
	//			To: api.RouteTargetReference{
	//				Name: serviceName,
	//			},
	//			Port: resolveRoutePort(portString),
	//		},
	return &r
	//oclient.RouteInterface().Create(r)
}

//func resolveRoutePort(portString string) *api.RoutePort {
//	if len(portString) == 0 {
//		return nil
//	}
//	var routePort intstr.IntOrString
//	integer, err := strconv.Atoi(portString)
//	if err != nil {
//		routePort = intstr.FromString(portString)
//	} else {
//		routePort = intstr.FromInt(integer)
//	}
//	return &api.RoutePort{
//		TargetPort: routePort,
//	}
//}
