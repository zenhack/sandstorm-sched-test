# You will likely want to adjust the paths below to point to the correct
# directories for both the sanstorm capnproto files and go-capnproto2;
# these are hardcoded for @zenhack's machine right now :/
capnp compile -ogo \
	-I /home/isd/src/pub/go.sandstorm/capnp/ \
	-I /home/isd/src/foreign/go-capnproto2/std/ \
	mixins.capnp
