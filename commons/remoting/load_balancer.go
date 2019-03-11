package remoting

type RemotingLoadBalancer interface {
	Selector(msg interface{}) string
}
