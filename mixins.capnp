@0x883486ed73c1aade;

using Go = import "/go.capnp";
using Grain = import "/grain.capnp";

$Go.package("main");
$Go.import("zenhack.net/go/sandstorm-sched-test");

interface AppPersistentCallback extends (Grain.ScheduledJob.Callback, Grain.AppPersistent) {}

struct AppObjectId {
  callbackName @0 :Text;
}
