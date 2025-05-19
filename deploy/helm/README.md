## Instructions:
Make this commands from .../deploy/helm folder

## Command for installation
```
helm install auth-service ./auth-service 
```

## If you want to override some value from cli (example)
```
helm install auth-service ./auth-service --set secret.KAFKA_URL="kafka.shared-services.svc.cluster.local:29092"
```

If you want to override through file - read helm docs