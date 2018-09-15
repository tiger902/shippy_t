// consignment-service/handler.go
package main

import (
	pb "github.com/shipper/consignment-service/proto/consignment"
	vesselProto "github.com/shipper/vessel-service/proto/vessel"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"log"
)

// Service should implement all of the methods to stisfy the service we defined in our protobuf definition. You can check the interface in the generated code itself for the exact method signatures etc to give you a better idea.
type service struct {
	session      *mgo.Session
	vesselClient vesselProto.VesselServiceClient
}

func (s *service) GetRepo() Repository {
	return &ConsignmentRepository{s.session.Clone()}
}

// CreateConsignment - we created jst one method on our service, which is a create method, which takes a context and a request as an argument, these are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {
	defer s.GetRepo().Close()

	// Here we cal a client instance of our vessel service with our consignment weight, and the amount of containers as the capacity value
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})
	log.Printf("Found vessel: $s", vesselResponse.Vessel.Name)
	if err != nil {
		return err
	}

	// We set the VesselId as the vessel we got back from our vessel service
	req.VesselId = vesselResponse.Vessel.Id

	// Save our consignment
	err = s.GetRepo().Create(req)
	if err != nil {
		return err
	}

	// Return matching the 'Response' message we created in our protobuf definition
	res.Created = true
	res.Consignment = req

	return nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	defer s.GetRepo().Close()

	consignments, err := s.GetRepo().GetAll()
	if err != nil {
		return nil
	}

	res.Consignments = consignments

	return nil
}