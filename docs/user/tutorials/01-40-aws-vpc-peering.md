# Create VPC peering in AWS

This tutorial explains how to create a VPC peering connection between a remote VPC network and Kyma in AWS. The tutorial
assumes that the Cloud Manager module is enabled in your Kyma cluster. Follow the steps from this tutorial to create a 
new VPC network, and VM, and assign required permissions to the provided Kyma account and role in your AWS account. If you want to
use the existing resources instead of creating new ones, adjust variable names accordingly and skip the steps that 
create those resources.

## Steps <!-- {docsify-ignore} -->

1.  Set default AWS CLI profile. If you haven't configured the profile it yet please go see AWS documentation https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-configure.html
    ```shell
    export AWS_PROFILE={PROFILE_NAME}
    export AWS_DEFAULT_REGION={REGION}
    ```
   
2.  Create trust policy document
    ```shell
    export PRINCIPAL_PROFILE_AWS_ACCOUNT_ID={TBD}
    export USER_NAME={TBD}
    cat > trust_policy.json <<- EOF
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "AWS": "arn:aws:iam::$PRINCIPAL_PROFILE_AWS_ACCOUNT_ID:user/$USER_NAME"
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }
    EOF
    ```
3. Create VpcPeeringRole and attach a trust policy to document.
    ```shell
    export AWS_ROLE_NAME={TBD}
    aws iam create-role --role-name $AWS_ROLE_NAME --assume-role-policy-document file://./trust_policy.json 
    ```
4.  Create policy document that will be used to create policy
    ```shell
    cat > accept_policy.json <<- EOF
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "ec2:AcceptVpcPeeringConnection",
                    "ec2:DescribeVpcs",
                    "ec2:DescribeVpcPeeringConnections",
                    "ec2:DescribeRouteTables",
                    "ec2:CreateRoute",
                    "ec2:CreateTags"
                ],
                "Resource": "*"
            }
        ]
    }
    EOF
    ```
   
5.  Creates a new managed policy for your Amazon Web Services account.
    ```shell
    aws iam create-policy --policy-name CloudManagerPeeringAccess --policy-document file://./accept_policy.json
    ```
6.  Attaches the specified managed policy to the specified IAM role.
    ```shell
    aws iam attach-role-policy --role-name $AWS_ROLE_NAME --policy-arn arn:aws:iam::$REMOTE_ACCOUNT_ID:policy/CloudManagerPeeringAccess
    ```
7.  Create a VPC and tag it with Kyma shoot name
    ```shell
    export CIDR_BLOCK=10.3.0.0/16
    export SHOOT_NAME=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}')
    export NODE_NETWORK=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.nodeNetwork}')
    export VPC_NAME=my-vpc
    export VPC_ID=$(aws ec2 create-vpc --cidr-block $CIDR_BLOCK --tag-specifications ResourceType=vpc,Tags=[{Key=$SHOOT_NAME,Value=""},{Key=Name,Value=$VPC_NAME}] --query Vpc.VpcId --output text)  
    ```
8.  Create subnet
    ```shell
    export SUBNET_ID=$(aws ec2 create-subnet --vpc-id $VPC_ID --cidr-block $CIDR_BLOCK --query Subnet.SubnetId --output text) 
    ```

9.  Run instance
    ```shell
    export INSTANCE_ID=$(aws ec2 run-instances --image-id ami-0c38b837cd80f13bb --instance-type t2.micro --subnet-id $SUBNET_ID --query "Instances[0].InstanceId" --output text)
    export IP_ADDRESS=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
    ```
10. Allow ICMP traffic from Kyma pods
    ```shell
     export SG_ID=$(aws ec2 describe-security-groups --filters Name=vpc-id,Values=$VPC_ID --query "SecurityGroups[0].GroupId" --output text) 
     aws ec2 authorize-security-group-ingress --group-id $SG_ID --ip-permissions IpProtocol=icmp,FromPort=-1,ToPort=-1,IpRanges="[{CidrIp=$NODE_NETWORK}]"
    ```

11. Create an AwsVpcPeering resource
    ```shell
    export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    kubectl apply -f - <<EOF
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: AwsVpcPeering
    metadata:
      name: peering-to-my-vpc
    spec:
      remoteAccount: $ACCOUNT_ID
      remoteRegion: $AWS_DEFAULT_REGION
      remoteVnet: $VPC_ID
    EOF
    ```

12. Wait for the AwsVpcPeering to be in the `Ready` state.

    ```shell
    kubectl wait --for=condition=Ready awsvpcpeering/peering-to-my-vpc --timeout=300s
    ```

    Once the newly created AwsVpcPeering is provisioned, you should see the following message:

    ```
    awsvpcpeering.cloud-resources.kyma-project.io/peering-to-my-vpc condition met
    ```

13. Create a namespace and export its value as an environment variable. Run:
    ```shell
    export NAMESPACE={NAMESPACE_NAME}
    kubectl create ns $NAMESPACE
    ```

14. Create a workload that pings the VM in the remote network.
    ```shell
    kubectl apply -n $NAMESPACE -f - <<EOF
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: awsvpcpeering-demo
    spec:
      selector:
        matchLabels:
          app: awsvpcpeering-demo
      template:
        metadata:
          labels:
            app: awsvpcpeering-demo
        spec:
          containers:
          - name: my-container
            resources:
              limits:
                memory: 512Mi
                cpu: "1"
              requests:
                memory: 256Mi
                cpu: "0.2"
            image: ubuntu
            command:
              - "/bin/bash"
              - "-c"
              - "--"
            args:
             - "apt update; apt install iputils-ping -y; ping -c 20 $IP_ADDRESS"
    EOF
    ```

This workload should print a sequence of 20 echo replies to stdout.

12. Print the logs of one of the workloads, run:
    ```shell
    kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=awsvpcpeering-demo -o=jsonpath='{.items[0].metadata.name}'`
    ```

    The command should print something like:
    ```
    ...
    PING 172.0.0.4 (172.0.0.4) 56(84) bytes of data.
    64 bytes from 172.0.0.4: icmp_seq=1 ttl=63 time=8.10 ms
    64 bytes from 172.0.0.4: icmp_seq=2 ttl=63 time=2.01 ms
    64 bytes from 172.0.0.4: icmp_seq=3 ttl=63 time=7.02 ms
    64 bytes from 172.0.0.4: icmp_seq=4 ttl=63 time=1.87 ms
    64 bytes from 172.0.0.4: icmp_seq=5 ttl=63 time=1.89 ms
    64 bytes from 172.0.0.4: icmp_seq=6 ttl=63 time=4.75 ms
    64 bytes from 172.0.0.4: icmp_seq=7 ttl=63 time=2.01 ms
    64 bytes from 172.0.0.4: icmp_seq=8 ttl=63 time=4.26 ms
    64 bytes from 172.0.0.4: icmp_seq=9 ttl=63 time=1.89 ms
    64 bytes from 172.0.0.4: icmp_seq=10 ttl=63 time=2.08 ms
    64 bytes from 172.0.0.4: icmp_seq=11 ttl=63 time=2.01 ms
    64 bytes from 172.0.0.4: icmp_seq=12 ttl=63 time=2.24 ms
    64 bytes from 172.0.0.4: icmp_seq=13 ttl=63 time=1.80 ms
    64 bytes from 172.0.0.4: icmp_seq=14 ttl=63 time=4.32 ms
    64 bytes from 172.0.0.4: icmp_seq=15 ttl=63 time=2.03 ms
    64 bytes from 172.0.0.4: icmp_seq=16 ttl=63 time=2.03 ms
    64 bytes from 172.0.0.4: icmp_seq=17 ttl=63 time=5.19 ms
    64 bytes from 172.0.0.4: icmp_seq=18 ttl=63 time=1.86 ms
    64 bytes from 172.0.0.4: icmp_seq=19 ttl=63 time=1.92 ms
    64 bytes from 172.0.0.4: icmp_seq=20 ttl=63 time=1.92 ms

    === 172.0.0.4 ping statistics ===
    20 packets transmitted, 20 received, 0% packet loss, time 19024ms
    rtt min/avg/max/mdev = 1.800/3.060/8.096/1.847 ms
    ...
    ```

13. Clean up Kubernetes resources

    * Remove the created workloads:
      ```shell
      kubectl delete -n $NAMESPACE deployment awsvpcpeering-demo
      ```

    * Remove the created `awsvpcpeering`:
      ```shell
      kubectl delete -n $NAMESPACE awsvpcpeering peering-to-my-vpc
      ```

    * Remove the created namespace:
      ```shell
      kubectl delete namespace $NAMESPACE
      ```

14. Clean up resources in your AWS account
    * Terminate instance
       ```shell
       aws ec2 terminate-instances --instance-ids $INSTANCE_ID
       ```
    * Delete subnet
      ```shell
      aws ec2 delete-subnet --subnet-id $SUBNET_ID
      ```
      
    * Delete VPC
      ```shell
      aws ec2 delete-vpc --vpc-id  $VPC_ID
      ```