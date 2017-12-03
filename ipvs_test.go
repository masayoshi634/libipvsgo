package libipvs

import (
	"testing"
)

func TestFlush(t *testing.T) {
	var err error
	ipvsHandler, err := NewIPVSHandler()
	if err != nil {
		t.Fatalf("Failed create IPVSHandler %s", err)
	}

	err = ipvsHandler.Flush()
	if err != nil {
		t.Fatalf("Failed Flush %s", err)
	}
}

func TestAddAndDeleteService(t *testing.T) {
	var err error

	vip := "127.1.1.1"
	vport := 8888
	protocol := "TCP"
	scheduler1 := "wlc"
	scheduler2 := "wrr"

	rip1 := "127.2.1.1"
	rip2 := "127.2.1.2"
	rport := 8888
	weight1 := 100
	weight2 := 200
	method := "DR"

	ipvsHandler, err := NewIPVSHandler()
	if err != nil {
		t.Fatalf("Failed create IPVSHandler %s", err)
	}

	// Add VIP
	err = ipvsHandler.AddService(vip, vport, protocol, scheduler1)
	if err != nil {
		t.Fatalf("Failed add service %s", err)
	}

	isRegisterd, err := ipvsHandler.IsRegisteredService(vip, vport, protocol)
	if err != nil {
		t.Fatalf("Failed IsRegisteredService %s", err)
	}
	if !isRegisterd {
		t.Fatalf("Failed IsRegisteredService %s", err)
	}

	// Update VIP
	err = ipvsHandler.UpdateService(vip, vport, "TCP", scheduler2)
	if err != nil {
		t.Fatalf("Failed update service %s", err)
	}

	service, err := ipvsHandler.GetService(vip, vport, protocol)
	if err != nil {
		t.Fatalf("Failed get service %s", err)
	}

	err = ipvsHandler.AddDestination(service, rip1, rport, weight1, method)
	err = ipvsHandler.AddDestination(service, rip2, rport, weight2, method)
	if err != nil {
		t.Fatalf("Failed add destination %s", err)
	}

	err = ipvsHandler.UpdateDestination(service, rip1, rport, weight2, method)
	err = ipvsHandler.UpdateDestination(service, rip2, rport, weight1, method)
	if err != nil {
		t.Fatalf("Failed update destination %s", err)
	}

	entries, err := ipvsHandler.GetAllEntry()
	if err != nil {
		t.Fatalf("Failed get all entries %s", err)
	}
	for _, entry := range entries {
		if entry.Service.Address != vip {
			t.Errorf("different address %s", entry.Service.Address)
		}
		if entry.Service.Protocol != protocol {
			t.Errorf("different protocol %s", entry.Service.Address)
		}
		for _, d := range entry.Destinations {
			if d.Method != method {
				t.Errorf("different method %s", d.Method)
			}
		}
	}

	err = ipvsHandler.DeleteService(vip, vport, protocol)
	if err != nil {
		t.Fatalf("Failed delete service %s", err)
	}
}
