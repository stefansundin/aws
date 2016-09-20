List ELBs with outdated security policy. [Check latest security policy](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-security-policy-options.html).
```shell
aws elb describe-load-balancers | jq '.LoadBalancerDescriptions | map(select(.Policies.OtherPolicies[0] != "ELBSecurityPolicy-2016-08")) | map(.LoadBalancerName)'
```
