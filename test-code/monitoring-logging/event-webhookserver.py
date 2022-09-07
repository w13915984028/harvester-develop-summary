import web
import json
import sys
from datetime import datetime

# http webhook server receives a bunch of k8s events per POST operation from fluentd, each EVENT has following format
#
#
# 
# {"verb":"ADDED","event":{"metadata":{"name":"virt-launcher-vm-ide-1-2p46t.170e43cb14ec7ede","namespace":"default","uid":"8aeb5ce9-4cfd-4364-b402-75c8179b7960","resourceVersion":"604569","creationTimestamp":"2022-08-24T11:17:32Z","managedFields":[{"manager":"kubelet","operation":"Update","apiVersion":"v1","time":"2022-08-24T11:17:32Z"}]},"involvedObject":{"kind":"Pod","namespace":"default","name":"virt-launcher-vm-ide-1-2p46t","uid":"8b718be5-8418-4d89-b21c-acd499b98db1","apiVersion":"v1","resourceVersion":"604401"},"reason":"SuccessfulMountVolume","message":"MapVolume.MapPodDevice succeeded for volume \"pvc-c73afadc-c6e6-46f4-b07f-a704acdbd75f\" globalMapPath \"/var/lib/kubelet/plugins/kubernetes.io/csi/volumeDevices/pvc-c73afadc-c6e6-46f4-b07f-a704acdbd75f/dev\"","source":{"component":"kubelet","host":"harv1"},"firstTimestamp":"2022-08-24T11:17:32Z","lastTimestamp":"2022-08-24T11:17:32Z","count":1,"type":"Normal","eventTime":null,"reportingComponent":"","reportingInstance":""}}

# this parse tries to print abstract information of k8s event
#
# -------event: 37------------
# event  : cattle-monitoring-system/rancher-monitoring-grafana-647fd577f4-gl72f.17128595f3c1895b
# message: Started container grafana
# time   : 2022-09-07T08:08:11Z manager:kubelet operation:Update
# object : kind:Pod cattle-monitoring-system/rancher-monitoring-grafana-647fd577f4-gl72f
#


# https://stackoverflow.com/questions/287871/how-do-i-print-colored-text-to-the-terminal
class bcolors:
  HEADER = '\033[95m'
  OKBLUE = '\033[94m'
  OKCYAN = '\033[96m'
  OKGREEN = '\033[92m'
  WARNING = '\033[93m'
  FAIL = '\033[91m'
  ENDC = '\033[0m'
  BOLD = '\033[1m'
  UNDERLINE = '\033[4m'

urls = ('/.*', 'hooks')
app = web.application(urls, globals())

class event_json_parser:
  def __init__(self, j_str, verbose_print):
    self.json_str = j_str
    self.verbose_print = verbose_print

  def parse(self):
    datas = self.json_str.split("\n")
    self.expect  = len(datas)
    self.good = 0
    self.bad = 0
    print('EVENT DATA RECEIVED len: '+str(len(self.json_str)) +" contains: " +str(self.expect) +"  " + str(datetime.now()))

    idx = 1
    for dd in datas:
        print("-------event: {}------------".format(idx))
        self.parse_event_json_str(dd)
        print('')
        idx +=1

    self.summary()

  def summary(self):
    print("expect:{}, good:{}, bad:{}".format(self.expect, self.good, self.bad))
    print('')

  def parse_event_json_str(self, dd):
    if self.verbose_print:
      print("event data: " + dd)

    if len(dd) < 2: # NULL end ? skip
      print("{}SKIP: this event, len:{} may be null, CHECK{}: {}".format(bcolors.WARNING, len(dd), bcolors.ENDC, dd))
      self.bad += 1
      return True

    try:
      jdata = json.loads(dd)
    except Exception as e:
      print("{}SKIP: try to load data as JSON, but error{}: {}".format( bcolors.WARNING, bcolors.ENDC, e))
      print("data:"+dd)
      self.bad += 1
      return True

    if "message" not in jdata:
      print("{}SKIP: the message is not in event, error{}: {}".format(bcolors.WARNING, bcolors.ENDC, dd))
      self.bad += 1
      return True

    msg = jdata["message"]
    try:
      jdata1 = json.loads(msg)
    except Exception as e:
      print("{}SKIP: try to load message as JSON, but error{}: {}".format( bcolors.WARNING, bcolors.ENDC, e))
      print("message:"+msg)
      self.bad += 1
      return True
      
    if "event" not in jdata1:
      print("{}SKIP: the message.event is not in message, error{} : {}".format( bcolors.WARNING, bcolors.ENDC, msg))
      self.bad += 1
      return True

    event = jdata1["event"] #message.event, a Dict

    if "metadata" not in event:
      print("{}SKIP: the message.event.metadata is not in event, error{}: {}".format( bcolors.WARNING, bcolors.ENDC, event))
      self.bad += 1
      return True

    meta = event["metadata"] #message.event.metadata

    try:
      print("event  : " + meta["namespace"] + "/" + meta["name"])

      if "message" not in event:
        print("message: ") # message may be NULL
      else:
        print("message: " + event["message"])

      if "managedFields" in meta:
        for x in meta["managedFields"]:
          print("time   : "+x["time"]+" manager:"+x["manager"]+" operation:"+x["operation"])

      if "involvedObject" in event:
        obj=event["involvedObject"]
        if "namespace" in obj:
          print("object : kind:" + obj["kind"] + " " + obj["namespace"] + "/" + obj["name"])
        else:
          print("object : kind:" + obj["kind"] + " " + "!!NO-NAMESPACE!!" + "/" + obj["name"])
      else:
        print("object : !!NO-OBJECT!!")
      self.good += 1

    except Exception as e:
      print("{}SKIP: try to parse an event, but error{}: {}; event:{}".format( bcolors.WARNING, bcolors.ENDC, e, event))
      self.bad += 1

    return True

class hooks:
    def POST(self):
        data = web.data().decode("utf-8")
        ejp = event_json_parser(data, len(sys.argv) > 1)
        ejp.parse()
        return 'OK'

if __name__ == '__main__':
  print("start a simple event webhook server at:" + str(datetime.now()))
  print("use export PORT=8090(e.g.) to set http server port number as 8090")
  # any additional param to open verbose print of event
  if len(sys.argv) > 1:
    print("verbose print is ON")
  else:
    print("verbose print is OFF")

  app.run()

