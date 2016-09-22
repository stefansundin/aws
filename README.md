List all TCP ELBs:
```shell
aws elb describe-load-balancers | jq '.LoadBalancerDescriptions | map(select(any(.ListenerDescriptions[]; .Listener.Protocol == "TCP"))) | map(.LoadBalancerName)'
```

List ELBs with an outdated security policy. [Check latest security policy](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-security-policy-options.html).
```shell
aws elb describe-load-balancers | jq '.LoadBalancerDescriptions | map(select(any(.ListenerDescriptions[]; .Listener.Protocol == "HTTPS"))) | map(select(.Policies.OtherPolicies[0] == "ELBSecurityPolicy-2015-05")) | map(.LoadBalancerName)'
```
Note that when updating to a more recent default policy, the name is not allowed to start with "ELBSecurityPolicy". That name can only be used by the default policy when the ELB is created. Annoying. So this command is a bit useless when updating a second time.
