using Go = import "/go.capnp";
@0xa69e0e7963207685;
$Go.package("service");
$Go.import("service");

interface Pinger {
	ping @0 (msg :Text) -> (msg :Text);
}