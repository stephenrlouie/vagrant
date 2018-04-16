import sys
import json

with open("/etc/kubernetes/admin.conf") as f:
    kubeconfig = f.read()

with open(sys.argv[1]) as fp:
    cluster_str = fp.read()
    cluster_data = json.loads(cluster_str)

cluster_data["metadata"]["annotations"]["Conf"] = kubeconfig

with open('/home/vagrant/my-cluster.json', 'w') as fp:
    json.dump(cluster_data, fp)


print "******** SUCCESSFULLY WROTE OPTIKON API POST JSON *********** "
