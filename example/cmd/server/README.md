# twirp example server #

This binary is an example twirp server. It's meant mostly to be read
and to be used in conjunction with the neighboring client binary.

When a request is made, the server will log the statsd messages it
would have sent, so you'll see stuff like this:

    -> % ./server
    incr twirp.total.requests: 1 @ 1.000000
    incr twirp.MakeHat.requests: 1 @ 1.000000
    incr twirp.total.responses: 1 @ 1.000000
    incr twirp.MakeHat.responses: 1 @ 1.000000
    incr twirp.status_codes.total.200: 1 @ 1.000000
    incr twirp.status_codes.MakeHat.200: 1 @ 1.000000
    time twirp.all_methods.response: 370.695µs @ 1.000000
    time twirp.MakeHat.response: 370.695µs @ 1.000000
    time twirp.status_codes.all_methods.200: 370.695µs @ 1.000000
    time twirp.status_codes.MakeHat.200: 370.695µs @ 1.000000
